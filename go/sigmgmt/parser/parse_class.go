// Copyright 2018 Anapaya Systems

package parser

import (
	"net"
	"strconv"

	"github.com/antlr/antlr4/runtime/Go/antlr"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/pktcls"
	"github.com/scionproto/scion/go/sigmgmt/parser/traffic_class"
)

type classListener struct {
	*traffic_class.BaseTrafficClassListener
	// condStack is a stack to store the Conds while the tree is parsed
	condStack []pktcls.Cond
	// countStack tracks the number of children conditions for each open parent condition
	// in the parser
	countStack []int
	err        error
}

func (l *classListener) popCond() pktcls.Cond {
	cond := l.condStack[len(l.condStack)-1]
	// Decrease condStack size and update indStack
	l.condStack = l.condStack[:len(l.condStack)-1]
	l.countStack = l.countStack[:len(l.countStack)-1]
	return cond
}

func (l *classListener) popConds() []pktcls.Cond {
	// Copy n cond into new slice
	n := l.countStack[len(l.countStack)-1]
	conds := make([]pktcls.Cond, n)
	copy(conds, l.condStack[len(l.condStack)-n:])
	// Decrease condStack size and update indStack
	l.condStack = l.condStack[:len(l.condStack)-n]
	l.countStack = l.countStack[:len(l.countStack)-1]
	return conds
}

func (l *classListener) pushCond(cond pktcls.Cond) {
	l.condStack = append(l.condStack, cond)
	l.countStack[len(l.countStack)-1]++
}

// incCondDep registers a new open parent condition on the condStack
func (l *classListener) incCondDep() {
	l.countStack = append(l.countStack, 0)
}

func (l *classListener) EnterTrafficClass(ctx *traffic_class.TrafficClassContext) {
	l.incCondDep()
}

func (l *classListener) EnterMatchDst(ctx *traffic_class.MatchDstContext) {
	// Push Selector as Predicate on stack and update the number of Conds on the stack
	var err error
	mdst := &pktcls.IPv4MatchDestination{}
	_, mdst.Net, err = net.ParseCIDR(ctx.GetStop().GetText())
	if err != nil {
		l.err = common.NewBasicError("CIDR parsing failed!", err, "cidr", ctx.GetStop().GetText())
	}
	l.pushCond(pktcls.NewCondIPv4(mdst))
}

func (l *classListener) EnterMatchSrc(ctx *traffic_class.MatchSrcContext) {
	// Push Selector as Predicate on stack and update the number of Conds on the stack
	var err error
	msrc := &pktcls.IPv4MatchSource{}
	_, msrc.Net, err = net.ParseCIDR(ctx.GetStop().GetText())
	if err != nil {
		l.err = common.NewBasicError("CIDR parsing failed!", err, "cidr", ctx.GetStop().GetText())
	}
	l.pushCond(pktcls.NewCondIPv4(msrc))
}

func (l *classListener) EnterMatchDSCP(ctx *traffic_class.MatchDSCPContext) {
	// Push Selector as Predicate on stack and update the number of Conds on the stack
	mdscp := &pktcls.IPv4MatchDSCP{}
	dscp, err := strconv.ParseUint(ctx.GetStop().GetText(), 16, 8)
	if err != nil {
		l.err = common.NewBasicError("DSCP parsing failed!", err, "dscp", ctx.GetStop().GetText())
	}
	mdscp.DSCP = uint8(dscp)
	l.pushCond(pktcls.NewCondIPv4(mdscp))
}

func (l *classListener) EnterMatchTOS(ctx *traffic_class.MatchTOSContext) {
	// Push Selector as Predicate on stack and update the number of Conds on the stack
	mtos := &pktcls.IPv4MatchToS{}
	tos, err := strconv.ParseUint(ctx.GetStop().GetText(), 16, 8)
	if err != nil {
		l.err = common.NewBasicError("TOS parsing failed!", err, "tos", ctx.GetStop().GetText())
	}
	mtos.TOS = uint8(tos)
	l.pushCond(pktcls.NewCondIPv4(mtos))
}

func (l *classListener) EnterCondCls(ctx *traffic_class.CondClsContext) {
	l.pushCond(NewCondClass(ctx.GetStop().GetText()))
}

func (l *classListener) EnterCondAny(ctx *traffic_class.CondAnyContext) {
	l.incCondDep()
}

func (l *classListener) ExitCondAny(ctx *traffic_class.CondAnyContext) {
	// Create CondAny from Conds and push it on stack
	conds := l.popConds()
	l.pushCond(pktcls.NewCondAnyOf(conds...))
}

func (l *classListener) EnterCondAll(ctx *traffic_class.CondAllContext) {
	l.incCondDep()
}

func (l *classListener) ExitCondAll(ctx *traffic_class.CondAllContext) {
	// Create CondAll from Conds and push it on stack
	conds := l.popConds()
	l.pushCond(pktcls.NewCondAllOf(conds...))
}

func (l *classListener) EnterCondNot(ctx *traffic_class.CondNotContext) {
	l.incCondDep()
}

func (l *classListener) ExitCondNot(ctx *traffic_class.CondNotContext) {
	// Create CondNot and push it on stack
	l.pushCond(pktcls.NewCondNot(l.popCond()))
}

func (l *classListener) EnterCondBool(ctx *traffic_class.CondBoolContext) {
	bool, err := strconv.ParseBool(ctx.GetStop().GetText())
	if err != nil {
		l.err = common.NewBasicError("CondBool parsing failed!", err, "bool", ctx.GetStop().GetText())
	}
	l.pushCond(pktcls.CondBool(bool))
}

// ValidateTrafficClass validates the structure of the class param
func ValidateTrafficClass(class string) error {
	p := buildTrafficClassParser(class)
	p.RemoveErrorListeners()
	errListener := &ErrorListener{errorType: "Parser"}
	p.AddErrorListener(errListener)
	// Walk the tree to validate the traffic class
	listener := &classListener{}
	antlr.ParseTreeWalkerDefault.Walk(listener, p.TrafficClass())
	if errListener.msg != "" {
		return common.NewBasicError("Parsing of traffic class failed:", nil,
			"err", errListener.msg)
	}
	if listener.err != nil {
		return listener.err
	}
	return nil
}

// BuildClassTree creates a Cond tree from the class param
func BuildClassTree(class string) (pktcls.Cond, error) {
	p := buildTrafficClassParser(class)
	p.RemoveErrorListeners()
	errListener := &ErrorListener{errorType: "Parser"}
	p.AddErrorListener(errListener)
	// Walk the tree and build the traffic class
	listener := &classListener{}
	antlr.ParseTreeWalkerDefault.Walk(listener, p.TrafficClass())
	if errListener.msg != "" {
		return nil, common.NewBasicError("Parsing of traffic class failed:", nil,
			"err", errListener.msg)
	}
	if listener.err != nil {
		return nil, listener.err
	}
	return listener.condStack[0], nil
}

func buildTrafficClassParser(class string) *traffic_class.TrafficClassParser {
	lexer := traffic_class.NewTrafficClassLexer(
		antlr.NewInputStream(class),
	)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(&ErrorListener{errorType: "Lexer"})
	parser := traffic_class.NewTrafficClassParser(
		antlr.NewCommonTokenStream(
			lexer,
			antlr.TokenDefaultChannel),
	)
	parser.BuildParseTrees = true
	return parser
}
