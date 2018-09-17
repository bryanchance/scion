// Code generated from PathPredicate.g4 by ANTLR 4.7.1. DO NOT EDIT.

package path_predicate // PathPredicate
import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = reflect.Copy
var _ = strconv.Itoa

var parserATN = []uint16{
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 3, 14, 59, 4,
	2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 4, 7, 9, 7, 3,
	2, 3, 2, 3, 2, 3, 2, 3, 2, 3, 2, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 7, 3, 26,
	10, 3, 12, 3, 14, 3, 29, 11, 3, 3, 3, 3, 3, 3, 4, 3, 4, 3, 4, 3, 4, 3,
	4, 7, 4, 38, 10, 4, 12, 4, 14, 4, 41, 11, 4, 3, 4, 3, 4, 3, 5, 3, 5, 3,
	5, 3, 5, 3, 5, 3, 6, 3, 6, 3, 6, 3, 6, 5, 6, 54, 10, 6, 3, 7, 3, 7, 3,
	7, 3, 7, 2, 2, 8, 2, 4, 6, 8, 10, 12, 2, 3, 4, 2, 9, 9, 11, 11, 2, 57,
	2, 14, 3, 2, 2, 2, 4, 20, 3, 2, 2, 2, 6, 32, 3, 2, 2, 2, 8, 44, 3, 2, 2,
	2, 10, 53, 3, 2, 2, 2, 12, 55, 3, 2, 2, 2, 14, 15, 7, 9, 2, 2, 15, 16,
	7, 3, 2, 2, 16, 17, 9, 2, 2, 2, 17, 18, 7, 4, 2, 2, 18, 19, 7, 9, 2, 2,
	19, 3, 3, 2, 2, 2, 20, 21, 7, 12, 2, 2, 21, 22, 7, 5, 2, 2, 22, 27, 5,
	10, 6, 2, 23, 24, 7, 6, 2, 2, 24, 26, 5, 10, 6, 2, 25, 23, 3, 2, 2, 2,
	26, 29, 3, 2, 2, 2, 27, 25, 3, 2, 2, 2, 27, 28, 3, 2, 2, 2, 28, 30, 3,
	2, 2, 2, 29, 27, 3, 2, 2, 2, 30, 31, 7, 7, 2, 2, 31, 5, 3, 2, 2, 2, 32,
	33, 7, 13, 2, 2, 33, 34, 7, 5, 2, 2, 34, 39, 5, 10, 6, 2, 35, 36, 7, 6,
	2, 2, 36, 38, 5, 10, 6, 2, 37, 35, 3, 2, 2, 2, 38, 41, 3, 2, 2, 2, 39,
	37, 3, 2, 2, 2, 39, 40, 3, 2, 2, 2, 40, 42, 3, 2, 2, 2, 41, 39, 3, 2, 2,
	2, 42, 43, 7, 7, 2, 2, 43, 7, 3, 2, 2, 2, 44, 45, 7, 14, 2, 2, 45, 46,
	7, 5, 2, 2, 46, 47, 5, 10, 6, 2, 47, 48, 7, 7, 2, 2, 48, 9, 3, 2, 2, 2,
	49, 54, 5, 6, 4, 2, 50, 54, 5, 4, 3, 2, 51, 54, 5, 8, 5, 2, 52, 54, 5,
	2, 2, 2, 53, 49, 3, 2, 2, 2, 53, 50, 3, 2, 2, 2, 53, 51, 3, 2, 2, 2, 53,
	52, 3, 2, 2, 2, 54, 11, 3, 2, 2, 2, 55, 56, 5, 10, 6, 2, 56, 57, 7, 2,
	2, 3, 57, 13, 3, 2, 2, 2, 5, 27, 39, 53,
}
var deserializer = antlr.NewATNDeserializer(nil)
var deserializedATN = deserializer.DeserializeFromUInt16(parserATN)

var literalNames = []string{
	"", "'-'", "'#'", "'('", "','", "')'",
}
var symbolicNames = []string{
	"", "", "", "", "", "", "WHITESPACE", "DIGITS", "HEX_DIGITS", "AS", "ANY",
	"ALL", "NOT",
}

var ruleNames = []string{
	"selector", "condAny", "condAll", "condNot", "cond", "pathPredicate",
}
var decisionToDFA = make([]*antlr.DFA, len(deserializedATN.DecisionToState))

func init() {
	for index, ds := range deserializedATN.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(ds, index)
	}
}

type PathPredicateParser struct {
	*antlr.BaseParser
}

func NewPathPredicateParser(input antlr.TokenStream) *PathPredicateParser {
	this := new(PathPredicateParser)

	this.BaseParser = antlr.NewBaseParser(input)

	this.Interpreter = antlr.NewParserATNSimulator(this, deserializedATN, decisionToDFA,
		antlr.NewPredictionContextCache())
	this.RuleNames = ruleNames
	this.LiteralNames = literalNames
	this.SymbolicNames = symbolicNames
	this.GrammarFileName = "PathPredicate.g4"

	return this
}

// PathPredicateParser tokens.
const (
	PathPredicateParserEOF        = antlr.TokenEOF
	PathPredicateParserT__0       = 1
	PathPredicateParserT__1       = 2
	PathPredicateParserT__2       = 3
	PathPredicateParserT__3       = 4
	PathPredicateParserT__4       = 5
	PathPredicateParserWHITESPACE = 6
	PathPredicateParserDIGITS     = 7
	PathPredicateParserHEX_DIGITS = 8
	PathPredicateParserAS         = 9
	PathPredicateParserANY        = 10
	PathPredicateParserALL        = 11
	PathPredicateParserNOT        = 12
)

// PathPredicateParser rules.
const (
	PathPredicateParserRULE_selector      = 0
	PathPredicateParserRULE_condAny       = 1
	PathPredicateParserRULE_condAll       = 2
	PathPredicateParserRULE_condNot       = 3
	PathPredicateParserRULE_cond          = 4
	PathPredicateParserRULE_pathPredicate = 5
)

// ISelectorContext is an interface to support dynamic dispatch.
type ISelectorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsSelectorContext differentiates from other interfaces.
	IsSelectorContext()
}

type SelectorContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySelectorContext() *SelectorContext {
	var p = new(SelectorContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = PathPredicateParserRULE_selector
	return p
}

func (*SelectorContext) IsSelectorContext() {}

func NewSelectorContext(parser antlr.Parser, parent antlr.ParserRuleContext,
	invokingState int) *SelectorContext {

	var p = new(SelectorContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = PathPredicateParserRULE_selector

	return p
}

func (s *SelectorContext) GetParser() antlr.Parser { return s.parser }

func (s *SelectorContext) AllDIGITS() []antlr.TerminalNode {
	return s.GetTokens(PathPredicateParserDIGITS)
}

func (s *SelectorContext) DIGITS(i int) antlr.TerminalNode {
	return s.GetToken(PathPredicateParserDIGITS, i)
}

func (s *SelectorContext) AS() antlr.TerminalNode {
	return s.GetToken(PathPredicateParserAS, 0)
}

func (s *SelectorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SelectorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SelectorContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(PathPredicateListener); ok {
		listenerT.EnterSelector(s)
	}
}

func (s *SelectorContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(PathPredicateListener); ok {
		listenerT.ExitSelector(s)
	}
}

func (p *PathPredicateParser) Selector() (localctx ISelectorContext) {
	localctx = NewSelectorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, PathPredicateParserRULE_selector)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(12)
		p.Match(PathPredicateParserDIGITS)
	}
	{
		p.SetState(13)
		p.Match(PathPredicateParserT__0)
	}
	{
		p.SetState(14)
		_la = p.GetTokenStream().LA(1)

		if !(_la == PathPredicateParserDIGITS || _la == PathPredicateParserAS) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}
	{
		p.SetState(15)
		p.Match(PathPredicateParserT__1)
	}
	{
		p.SetState(16)
		p.Match(PathPredicateParserDIGITS)
	}

	return localctx
}

// ICondAnyContext is an interface to support dynamic dispatch.
type ICondAnyContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsCondAnyContext differentiates from other interfaces.
	IsCondAnyContext()
}

type CondAnyContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCondAnyContext() *CondAnyContext {
	var p = new(CondAnyContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = PathPredicateParserRULE_condAny
	return p
}

func (*CondAnyContext) IsCondAnyContext() {}

func NewCondAnyContext(parser antlr.Parser, parent antlr.ParserRuleContext,
	invokingState int) *CondAnyContext {

	var p = new(CondAnyContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = PathPredicateParserRULE_condAny

	return p
}

func (s *CondAnyContext) GetParser() antlr.Parser { return s.parser }

func (s *CondAnyContext) ANY() antlr.TerminalNode {
	return s.GetToken(PathPredicateParserANY, 0)
}

func (s *CondAnyContext) AllCond() []ICondContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*ICondContext)(nil)).Elem())
	var tst = make([]ICondContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(ICondContext)
		}
	}

	return tst
}

func (s *CondAnyContext) Cond(i int) ICondContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ICondContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(ICondContext)
}

func (s *CondAnyContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CondAnyContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CondAnyContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(PathPredicateListener); ok {
		listenerT.EnterCondAny(s)
	}
}

func (s *CondAnyContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(PathPredicateListener); ok {
		listenerT.ExitCondAny(s)
	}
}

func (p *PathPredicateParser) CondAny() (localctx ICondAnyContext) {
	localctx = NewCondAnyContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, PathPredicateParserRULE_condAny)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(18)
		p.Match(PathPredicateParserANY)
	}
	{
		p.SetState(19)
		p.Match(PathPredicateParserT__2)
	}
	{
		p.SetState(20)
		p.Cond()
	}
	p.SetState(25)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for _la == PathPredicateParserT__3 {
		{
			p.SetState(21)
			p.Match(PathPredicateParserT__3)
		}
		{
			p.SetState(22)
			p.Cond()
		}

		p.SetState(27)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(28)
		p.Match(PathPredicateParserT__4)
	}

	return localctx
}

// ICondAllContext is an interface to support dynamic dispatch.
type ICondAllContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsCondAllContext differentiates from other interfaces.
	IsCondAllContext()
}

type CondAllContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCondAllContext() *CondAllContext {
	var p = new(CondAllContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = PathPredicateParserRULE_condAll
	return p
}

func (*CondAllContext) IsCondAllContext() {}

func NewCondAllContext(parser antlr.Parser, parent antlr.ParserRuleContext,
	invokingState int) *CondAllContext {

	var p = new(CondAllContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = PathPredicateParserRULE_condAll

	return p
}

func (s *CondAllContext) GetParser() antlr.Parser { return s.parser }

func (s *CondAllContext) ALL() antlr.TerminalNode {
	return s.GetToken(PathPredicateParserALL, 0)
}

func (s *CondAllContext) AllCond() []ICondContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*ICondContext)(nil)).Elem())
	var tst = make([]ICondContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(ICondContext)
		}
	}

	return tst
}

func (s *CondAllContext) Cond(i int) ICondContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ICondContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(ICondContext)
}

func (s *CondAllContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CondAllContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CondAllContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(PathPredicateListener); ok {
		listenerT.EnterCondAll(s)
	}
}

func (s *CondAllContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(PathPredicateListener); ok {
		listenerT.ExitCondAll(s)
	}
}

func (p *PathPredicateParser) CondAll() (localctx ICondAllContext) {
	localctx = NewCondAllContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, PathPredicateParserRULE_condAll)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(30)
		p.Match(PathPredicateParserALL)
	}
	{
		p.SetState(31)
		p.Match(PathPredicateParserT__2)
	}
	{
		p.SetState(32)
		p.Cond()
	}
	p.SetState(37)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for _la == PathPredicateParserT__3 {
		{
			p.SetState(33)
			p.Match(PathPredicateParserT__3)
		}
		{
			p.SetState(34)
			p.Cond()
		}

		p.SetState(39)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(40)
		p.Match(PathPredicateParserT__4)
	}

	return localctx
}

// ICondNotContext is an interface to support dynamic dispatch.
type ICondNotContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsCondNotContext differentiates from other interfaces.
	IsCondNotContext()
}

type CondNotContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCondNotContext() *CondNotContext {
	var p = new(CondNotContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = PathPredicateParserRULE_condNot
	return p
}

func (*CondNotContext) IsCondNotContext() {}

func NewCondNotContext(parser antlr.Parser, parent antlr.ParserRuleContext,
	invokingState int) *CondNotContext {

	var p = new(CondNotContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = PathPredicateParserRULE_condNot

	return p
}

func (s *CondNotContext) GetParser() antlr.Parser { return s.parser }

func (s *CondNotContext) NOT() antlr.TerminalNode {
	return s.GetToken(PathPredicateParserNOT, 0)
}

func (s *CondNotContext) Cond() ICondContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ICondContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ICondContext)
}

func (s *CondNotContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CondNotContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CondNotContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(PathPredicateListener); ok {
		listenerT.EnterCondNot(s)
	}
}

func (s *CondNotContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(PathPredicateListener); ok {
		listenerT.ExitCondNot(s)
	}
}

func (p *PathPredicateParser) CondNot() (localctx ICondNotContext) {
	localctx = NewCondNotContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, PathPredicateParserRULE_condNot)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(42)
		p.Match(PathPredicateParserNOT)
	}
	{
		p.SetState(43)
		p.Match(PathPredicateParserT__2)
	}
	{
		p.SetState(44)
		p.Cond()
	}
	{
		p.SetState(45)
		p.Match(PathPredicateParserT__4)
	}

	return localctx
}

// ICondContext is an interface to support dynamic dispatch.
type ICondContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsCondContext differentiates from other interfaces.
	IsCondContext()
}

type CondContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCondContext() *CondContext {
	var p = new(CondContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = PathPredicateParserRULE_cond
	return p
}

func (*CondContext) IsCondContext() {}

func NewCondContext(parser antlr.Parser, parent antlr.ParserRuleContext,
	invokingState int) *CondContext {

	var p = new(CondContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = PathPredicateParserRULE_cond

	return p
}

func (s *CondContext) GetParser() antlr.Parser { return s.parser }

func (s *CondContext) CondAll() ICondAllContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ICondAllContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ICondAllContext)
}

func (s *CondContext) CondAny() ICondAnyContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ICondAnyContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ICondAnyContext)
}

func (s *CondContext) CondNot() ICondNotContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ICondNotContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ICondNotContext)
}

func (s *CondContext) Selector() ISelectorContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISelectorContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISelectorContext)
}

func (s *CondContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CondContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CondContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(PathPredicateListener); ok {
		listenerT.EnterCond(s)
	}
}

func (s *CondContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(PathPredicateListener); ok {
		listenerT.ExitCond(s)
	}
}

func (p *PathPredicateParser) Cond() (localctx ICondContext) {
	localctx = NewCondContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, PathPredicateParserRULE_cond)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.SetState(51)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case PathPredicateParserALL:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(47)
			p.CondAll()
		}

	case PathPredicateParserANY:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(48)
			p.CondAny()
		}

	case PathPredicateParserNOT:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(49)
			p.CondNot()
		}

	case PathPredicateParserDIGITS:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(50)
			p.Selector()
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}

	return localctx
}

// IPathPredicateContext is an interface to support dynamic dispatch.
type IPathPredicateContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsPathPredicateContext differentiates from other interfaces.
	IsPathPredicateContext()
}

type PathPredicateContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPathPredicateContext() *PathPredicateContext {
	var p = new(PathPredicateContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = PathPredicateParserRULE_pathPredicate
	return p
}

func (*PathPredicateContext) IsPathPredicateContext() {}

func NewPathPredicateContext(parser antlr.Parser, parent antlr.ParserRuleContext,
	invokingState int) *PathPredicateContext {

	var p = new(PathPredicateContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = PathPredicateParserRULE_pathPredicate

	return p
}

func (s *PathPredicateContext) GetParser() antlr.Parser { return s.parser }

func (s *PathPredicateContext) Cond() ICondContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ICondContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ICondContext)
}

func (s *PathPredicateContext) EOF() antlr.TerminalNode {
	return s.GetToken(PathPredicateParserEOF, 0)
}

func (s *PathPredicateContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PathPredicateContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PathPredicateContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(PathPredicateListener); ok {
		listenerT.EnterPathPredicate(s)
	}
}

func (s *PathPredicateContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(PathPredicateListener); ok {
		listenerT.ExitPathPredicate(s)
	}
}

func (p *PathPredicateParser) PathPredicate() (localctx IPathPredicateContext) {
	localctx = NewPathPredicateContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, PathPredicateParserRULE_pathPredicate)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(53)
		p.Cond()
	}
	{
		p.SetState(54)
		p.Match(PathPredicateParserEOF)
	}

	return localctx
}
