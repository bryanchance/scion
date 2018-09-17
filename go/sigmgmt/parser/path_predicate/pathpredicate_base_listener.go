// Code generated from PathPredicate.g4 by ANTLR 4.7.1. DO NOT EDIT.

package path_predicate // PathPredicate
import "github.com/antlr/antlr4/runtime/Go/antlr"

// BasePathPredicateListener is a complete listener for a parse tree produced
// by PathPredicateParser.
type BasePathPredicateListener struct{}

var _ PathPredicateListener = &BasePathPredicateListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BasePathPredicateListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BasePathPredicateListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BasePathPredicateListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BasePathPredicateListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterSelector is called when production selector is entered.
func (s *BasePathPredicateListener) EnterSelector(ctx *SelectorContext) {}

// ExitSelector is called when production selector is exited.
func (s *BasePathPredicateListener) ExitSelector(ctx *SelectorContext) {}

// EnterCondAny is called when production condAny is entered.
func (s *BasePathPredicateListener) EnterCondAny(ctx *CondAnyContext) {}

// ExitCondAny is called when production condAny is exited.
func (s *BasePathPredicateListener) ExitCondAny(ctx *CondAnyContext) {}

// EnterCondAll is called when production condAll is entered.
func (s *BasePathPredicateListener) EnterCondAll(ctx *CondAllContext) {}

// ExitCondAll is called when production condAll is exited.
func (s *BasePathPredicateListener) ExitCondAll(ctx *CondAllContext) {}

// EnterCondNot is called when production condNot is entered.
func (s *BasePathPredicateListener) EnterCondNot(ctx *CondNotContext) {}

// ExitCondNot is called when production condNot is exited.
func (s *BasePathPredicateListener) ExitCondNot(ctx *CondNotContext) {}

// EnterCond is called when production cond is entered.
func (s *BasePathPredicateListener) EnterCond(ctx *CondContext) {}

// ExitCond is called when production cond is exited.
func (s *BasePathPredicateListener) ExitCond(ctx *CondContext) {}

// EnterPathPredicate is called when production pathPredicate is entered.
func (s *BasePathPredicateListener) EnterPathPredicate(ctx *PathPredicateContext) {}

// ExitPathPredicate is called when production pathPredicate is exited.
func (s *BasePathPredicateListener) ExitPathPredicate(ctx *PathPredicateContext) {}
