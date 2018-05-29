// Copyright 2018 Anapaya Systems

package parser

import (
	"github.com/antlr/antlr4/runtime/Go/antlr"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/pktcls"
	"github.com/scionproto/scion/go/lib/spath/spathmeta"
)

type ErrorListener struct {
	*antlr.DefaultErrorListener
	msg string
}

func (c *ErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line,
	column int, msg string, e antlr.RecognitionException) {
	c.msg = msg
}

type predicateListener struct {
	*BasePathPredicateListener
	// condStack is a stack to store the Conds while the tree is parsed
	condStack []pktcls.Cond
	// countStack tracks the number of children conditions for each open parent condition
	// in the parser
	countStack []int
}

func (l *predicateListener) popCond() pktcls.Cond {
	cond := l.condStack[len(l.condStack)-1]
	// Decrease condStack size and update indStack
	l.condStack = l.condStack[:len(l.condStack)-1]
	l.countStack = l.countStack[:len(l.countStack)-1]
	return cond
}

func (l *predicateListener) popConds() []pktcls.Cond {
	// Copy n cond into new slice
	n := l.countStack[len(l.countStack)-1]
	conds := make([]pktcls.Cond, n)
	copy(conds, l.condStack[len(l.condStack)-n:])
	// Decrease condStack size and update indStack
	l.condStack = l.condStack[:len(l.condStack)-n]
	l.countStack = l.countStack[:len(l.countStack)-1]
	return conds
}

func (l *predicateListener) pushCond(cond pktcls.Cond) {
	l.condStack = append(l.condStack, cond)
	l.countStack[len(l.countStack)-1]++
}

// incCondDep registers a new open parent condition on the condStack
func (l *predicateListener) incCondDep() {
	l.countStack = append(l.countStack, 0)
}

func (l *predicateListener) EnterPathPredicate(ctx *PathPredicateContext) {
	l.incCondDep()
}

func (l *predicateListener) ExitSelector(ctx *SelectorContext) {
	// Push selector as predicate
	pp, _ := spathmeta.NewPathPredicate(ctx.GetText())
	l.pushCond(pktcls.NewCondPathPredicate(pp))
}

func (l *predicateListener) EnterCondAny(ctx *CondAnyContext) {
	l.incCondDep()
}

func (l *predicateListener) ExitCondAny(ctx *CondAnyContext) {
	// Create CondAny from Conds and push it on stack
	conds := l.popConds()
	l.pushCond(pktcls.NewCondAnyOf(conds...))
}

func (l *predicateListener) EnterCondAll(ctx *CondAllContext) {
	l.incCondDep()
}

func (l *predicateListener) ExitCondAll(ctx *CondAllContext) {
	// Create CondAll from Conds and push it on stack
	conds := l.popConds()
	l.pushCond(pktcls.NewCondAllOf(conds...))
}

func (l *predicateListener) EnterCondNot(ctx *CondNotContext) {
	l.incCondDep()
}

func (l *predicateListener) ExitCondNot(ctx *CondNotContext) {
	// Create CondNot and push it on stack
	l.pushCond(pktcls.NewCondNot(l.popCond()))
}

// ValidatePredicate validates the structure of the predicate param
func ValidatePredicate(predicate string) error {
	p := buildParser(predicate)
	errListener := &ErrorListener{}
	p.AddErrorListener(errListener)
	// Walk the tree to validate the predicate
	antlr.ParseTreeWalkerDefault.Walk(&predicateListener{}, p.PathPredicate())
	if errListener.msg != "" {
		return common.NewBasicError("Parsing of path predicate failed:", nil,
			"err", errListener.msg)
	}
	return nil
}

// BuildPredicateTree creates a Cond tree from the predicate param
func BuildPredicateTree(predicate string) (pktcls.Cond, error) {
	p := buildParser(predicate)
	errListener := &ErrorListener{}
	p.AddErrorListener(errListener)
	// Walk the tree and build the path predicate
	listener := &predicateListener{}
	antlr.ParseTreeWalkerDefault.Walk(listener, p.PathPredicate())
	if errListener.msg != "" {
		return nil, common.NewBasicError("Parsing of path predicate failed:", nil,
			"err", errListener.msg)
	}
	return listener.popCond(), nil
}

func buildParser(predicate string) *PathPredicateParser {
	parser := NewPathPredicateParser(
		antlr.NewCommonTokenStream(
			NewPathPredicateLexer(
				antlr.NewInputStream(predicate),
			),
			antlr.TokenDefaultChannel),
	)
	parser.BuildParseTrees = true
	return parser
}
