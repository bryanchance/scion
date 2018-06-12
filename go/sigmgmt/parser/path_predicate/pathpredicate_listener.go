// Code generated from PathPredicate.g4 by ANTLR 4.7.1. DO NOT EDIT.

package path_predicate // PathPredicate
import "github.com/antlr/antlr4/runtime/Go/antlr"

// PathPredicateListener is a complete listener for a parse tree produced by PathPredicateParser.
type PathPredicateListener interface {
	antlr.ParseTreeListener

	// EnterSelector is called when entering the selector production.
	EnterSelector(c *SelectorContext)

	// EnterCondAny is called when entering the condAny production.
	EnterCondAny(c *CondAnyContext)

	// EnterCondAll is called when entering the condAll production.
	EnterCondAll(c *CondAllContext)

	// EnterCondNot is called when entering the condNot production.
	EnterCondNot(c *CondNotContext)

	// EnterCond is called when entering the cond production.
	EnterCond(c *CondContext)

	// EnterPathPredicate is called when entering the pathPredicate production.
	EnterPathPredicate(c *PathPredicateContext)

	// ExitSelector is called when exiting the selector production.
	ExitSelector(c *SelectorContext)

	// ExitCondAny is called when exiting the condAny production.
	ExitCondAny(c *CondAnyContext)

	// ExitCondAll is called when exiting the condAll production.
	ExitCondAll(c *CondAllContext)

	// ExitCondNot is called when exiting the condNot production.
	ExitCondNot(c *CondNotContext)

	// ExitCond is called when exiting the cond production.
	ExitCond(c *CondContext)

	// ExitPathPredicate is called when exiting the pathPredicate production.
	ExitPathPredicate(c *PathPredicateContext)
}
