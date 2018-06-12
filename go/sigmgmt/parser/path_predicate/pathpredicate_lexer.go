// Code generated from PathPredicate.g4 by ANTLR 4.7.1. DO NOT EDIT.

package path_predicate

import (
	"fmt"
	"unicode"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = unicode.IsLetter

var serializedLexerAtn = []uint16{
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 2, 14, 89, 8,
	1, 4, 2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 4, 7, 9,
	7, 4, 8, 9, 8, 4, 9, 9, 9, 4, 10, 9, 10, 4, 11, 9, 11, 4, 12, 9, 12, 4,
	13, 9, 13, 3, 2, 3, 2, 3, 3, 3, 3, 3, 4, 3, 4, 3, 5, 3, 5, 3, 6, 3, 6,
	3, 7, 6, 7, 39, 10, 7, 13, 7, 14, 7, 40, 3, 7, 3, 7, 3, 8, 3, 8, 3, 8,
	7, 8, 48, 10, 8, 12, 8, 14, 8, 51, 11, 8, 5, 8, 53, 10, 8, 3, 9, 6, 9,
	56, 10, 9, 13, 9, 14, 9, 57, 3, 10, 3, 10, 3, 10, 3, 10, 3, 10, 3, 10,
	3, 11, 3, 11, 3, 11, 3, 11, 3, 11, 3, 11, 5, 11, 72, 10, 11, 3, 12, 3,
	12, 3, 12, 3, 12, 3, 12, 3, 12, 5, 12, 80, 10, 12, 3, 13, 3, 13, 3, 13,
	3, 13, 3, 13, 3, 13, 5, 13, 88, 10, 13, 2, 2, 14, 3, 3, 5, 4, 7, 5, 9,
	6, 11, 7, 13, 8, 15, 9, 17, 10, 19, 11, 21, 12, 23, 13, 25, 14, 3, 2, 6,
	5, 2, 11, 12, 15, 15, 34, 34, 3, 2, 51, 59, 3, 2, 50, 59, 5, 2, 50, 59,
	67, 72, 99, 104, 2, 95, 2, 3, 3, 2, 2, 2, 2, 5, 3, 2, 2, 2, 2, 7, 3, 2,
	2, 2, 2, 9, 3, 2, 2, 2, 2, 11, 3, 2, 2, 2, 2, 13, 3, 2, 2, 2, 2, 15, 3,
	2, 2, 2, 2, 17, 3, 2, 2, 2, 2, 19, 3, 2, 2, 2, 2, 21, 3, 2, 2, 2, 2, 23,
	3, 2, 2, 2, 2, 25, 3, 2, 2, 2, 3, 27, 3, 2, 2, 2, 5, 29, 3, 2, 2, 2, 7,
	31, 3, 2, 2, 2, 9, 33, 3, 2, 2, 2, 11, 35, 3, 2, 2, 2, 13, 38, 3, 2, 2,
	2, 15, 52, 3, 2, 2, 2, 17, 55, 3, 2, 2, 2, 19, 59, 3, 2, 2, 2, 21, 71,
	3, 2, 2, 2, 23, 79, 3, 2, 2, 2, 25, 87, 3, 2, 2, 2, 27, 28, 7, 47, 2, 2,
	28, 4, 3, 2, 2, 2, 29, 30, 7, 37, 2, 2, 30, 6, 3, 2, 2, 2, 31, 32, 7, 42,
	2, 2, 32, 8, 3, 2, 2, 2, 33, 34, 7, 46, 2, 2, 34, 10, 3, 2, 2, 2, 35, 36,
	7, 43, 2, 2, 36, 12, 3, 2, 2, 2, 37, 39, 9, 2, 2, 2, 38, 37, 3, 2, 2, 2,
	39, 40, 3, 2, 2, 2, 40, 38, 3, 2, 2, 2, 40, 41, 3, 2, 2, 2, 41, 42, 3,
	2, 2, 2, 42, 43, 8, 7, 2, 2, 43, 14, 3, 2, 2, 2, 44, 53, 7, 50, 2, 2, 45,
	49, 9, 3, 2, 2, 46, 48, 9, 4, 2, 2, 47, 46, 3, 2, 2, 2, 48, 51, 3, 2, 2,
	2, 49, 47, 3, 2, 2, 2, 49, 50, 3, 2, 2, 2, 50, 53, 3, 2, 2, 2, 51, 49,
	3, 2, 2, 2, 52, 44, 3, 2, 2, 2, 52, 45, 3, 2, 2, 2, 53, 16, 3, 2, 2, 2,
	54, 56, 9, 5, 2, 2, 55, 54, 3, 2, 2, 2, 56, 57, 3, 2, 2, 2, 57, 55, 3,
	2, 2, 2, 57, 58, 3, 2, 2, 2, 58, 18, 3, 2, 2, 2, 59, 60, 5, 17, 9, 2, 60,
	61, 7, 60, 2, 2, 61, 62, 5, 17, 9, 2, 62, 63, 7, 60, 2, 2, 63, 64, 5, 17,
	9, 2, 64, 20, 3, 2, 2, 2, 65, 66, 7, 67, 2, 2, 66, 67, 7, 80, 2, 2, 67,
	72, 7, 91, 2, 2, 68, 69, 7, 99, 2, 2, 69, 70, 7, 112, 2, 2, 70, 72, 7,
	123, 2, 2, 71, 65, 3, 2, 2, 2, 71, 68, 3, 2, 2, 2, 72, 22, 3, 2, 2, 2,
	73, 74, 7, 67, 2, 2, 74, 75, 7, 78, 2, 2, 75, 80, 7, 78, 2, 2, 76, 77,
	7, 99, 2, 2, 77, 78, 7, 110, 2, 2, 78, 80, 7, 110, 2, 2, 79, 73, 3, 2,
	2, 2, 79, 76, 3, 2, 2, 2, 80, 24, 3, 2, 2, 2, 81, 82, 7, 80, 2, 2, 82,
	83, 7, 81, 2, 2, 83, 88, 7, 86, 2, 2, 84, 85, 7, 112, 2, 2, 85, 86, 7,
	113, 2, 2, 86, 88, 7, 118, 2, 2, 87, 81, 3, 2, 2, 2, 87, 84, 3, 2, 2, 2,
	88, 26, 3, 2, 2, 2, 11, 2, 40, 49, 52, 55, 57, 71, 79, 87, 3, 8, 2, 2,
}

var lexerDeserializer = antlr.NewATNDeserializer(nil)
var lexerAtn = lexerDeserializer.DeserializeFromUInt16(serializedLexerAtn)

var lexerChannelNames = []string{
	"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
}

var lexerModeNames = []string{
	"DEFAULT_MODE",
}

var lexerLiteralNames = []string{
	"", "'-'", "'#'", "'('", "','", "')'",
}

var lexerSymbolicNames = []string{
	"", "", "", "", "", "", "WHITESPACE", "DIGITS", "HEX_DIGITS", "AS", "ANY",
	"ALL", "NOT",
}

var lexerRuleNames = []string{
	"T__0", "T__1", "T__2", "T__3", "T__4", "WHITESPACE", "DIGITS", "HEX_DIGITS",
	"AS", "ANY", "ALL", "NOT",
}

type PathPredicateLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var lexerDecisionToDFA = make([]*antlr.DFA, len(lexerAtn.DecisionToState))

func init() {
	for index, ds := range lexerAtn.DecisionToState {
		lexerDecisionToDFA[index] = antlr.NewDFA(ds, index)
	}
}

func NewPathPredicateLexer(input antlr.CharStream) *PathPredicateLexer {

	l := new(PathPredicateLexer)

	l.BaseLexer = antlr.NewBaseLexer(input)
	l.Interpreter = antlr.NewLexerATNSimulator(l, lexerAtn, lexerDecisionToDFA, antlr.NewPredictionContextCache())

	l.channelNames = lexerChannelNames
	l.modeNames = lexerModeNames
	l.RuleNames = lexerRuleNames
	l.LiteralNames = lexerLiteralNames
	l.SymbolicNames = lexerSymbolicNames
	l.GrammarFileName = "PathPredicate.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// PathPredicateLexer tokens.
const (
	PathPredicateLexerT__0       = 1
	PathPredicateLexerT__1       = 2
	PathPredicateLexerT__2       = 3
	PathPredicateLexerT__3       = 4
	PathPredicateLexerT__4       = 5
	PathPredicateLexerWHITESPACE = 6
	PathPredicateLexerDIGITS     = 7
	PathPredicateLexerHEX_DIGITS = 8
	PathPredicateLexerAS         = 9
	PathPredicateLexerANY        = 10
	PathPredicateLexerALL        = 11
	PathPredicateLexerNOT        = 12
)
