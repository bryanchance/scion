// Code generated from TrafficClass.g4 by ANTLR 4.7.1. DO NOT EDIT.

package traffic_class

import (
	"fmt"
	"unicode"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = unicode.IsLetter

var serializedLexerAtn = []uint16{
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 2, 19, 144,
	8, 1, 4, 2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 4, 7,
	9, 7, 4, 8, 9, 8, 4, 9, 9, 9, 4, 10, 9, 10, 4, 11, 9, 11, 4, 12, 9, 12,
	4, 13, 9, 13, 4, 14, 9, 14, 4, 15, 9, 15, 4, 16, 9, 16, 4, 17, 9, 17, 4,
	18, 9, 18, 3, 2, 3, 2, 3, 3, 3, 3, 3, 3, 3, 3, 3, 4, 3, 4, 3, 4, 3, 4,
	3, 4, 3, 5, 3, 5, 3, 6, 3, 6, 3, 7, 3, 7, 3, 8, 6, 8, 56, 10, 8, 13, 8,
	14, 8, 57, 3, 8, 3, 8, 3, 9, 3, 9, 3, 9, 7, 9, 65, 10, 9, 12, 9, 14, 9,
	68, 11, 9, 5, 9, 70, 10, 9, 3, 10, 6, 10, 73, 10, 10, 13, 10, 14, 10, 74,
	3, 11, 3, 11, 3, 11, 3, 11, 3, 11, 3, 11, 3, 11, 3, 11, 3, 11, 3, 11, 3,
	12, 3, 12, 3, 12, 3, 12, 3, 12, 3, 12, 5, 12, 93, 10, 12, 3, 13, 3, 13,
	3, 13, 3, 13, 3, 13, 3, 13, 5, 13, 101, 10, 13, 3, 14, 3, 14, 3, 14, 3,
	14, 3, 14, 3, 14, 5, 14, 109, 10, 14, 3, 15, 3, 15, 3, 15, 3, 15, 3, 15,
	3, 15, 5, 15, 117, 10, 15, 3, 16, 3, 16, 3, 16, 3, 16, 3, 16, 3, 16, 5,
	16, 125, 10, 16, 3, 17, 3, 17, 3, 17, 3, 17, 3, 17, 3, 17, 3, 17, 3, 17,
	5, 17, 135, 10, 17, 3, 18, 3, 18, 3, 18, 3, 18, 3, 18, 3, 18, 5, 18, 143,
	10, 18, 2, 2, 19, 3, 3, 5, 4, 7, 5, 9, 6, 11, 7, 13, 8, 15, 9, 17, 10,
	19, 11, 21, 12, 23, 13, 25, 14, 27, 15, 29, 16, 31, 17, 33, 18, 35, 19,
	3, 2, 6, 5, 2, 11, 12, 15, 15, 34, 34, 3, 2, 51, 59, 3, 2, 50, 59, 5, 2,
	50, 59, 67, 72, 99, 104, 2, 154, 2, 3, 3, 2, 2, 2, 2, 5, 3, 2, 2, 2, 2,
	7, 3, 2, 2, 2, 2, 9, 3, 2, 2, 2, 2, 11, 3, 2, 2, 2, 2, 13, 3, 2, 2, 2,
	2, 15, 3, 2, 2, 2, 2, 17, 3, 2, 2, 2, 2, 19, 3, 2, 2, 2, 2, 21, 3, 2, 2,
	2, 2, 23, 3, 2, 2, 2, 2, 25, 3, 2, 2, 2, 2, 27, 3, 2, 2, 2, 2, 29, 3, 2,
	2, 2, 2, 31, 3, 2, 2, 2, 2, 33, 3, 2, 2, 2, 2, 35, 3, 2, 2, 2, 3, 37, 3,
	2, 2, 2, 5, 39, 3, 2, 2, 2, 7, 43, 3, 2, 2, 2, 9, 48, 3, 2, 2, 2, 11, 50,
	3, 2, 2, 2, 13, 52, 3, 2, 2, 2, 15, 55, 3, 2, 2, 2, 17, 69, 3, 2, 2, 2,
	19, 72, 3, 2, 2, 2, 21, 76, 3, 2, 2, 2, 23, 92, 3, 2, 2, 2, 25, 100, 3,
	2, 2, 2, 27, 108, 3, 2, 2, 2, 29, 116, 3, 2, 2, 2, 31, 124, 3, 2, 2, 2,
	33, 134, 3, 2, 2, 2, 35, 142, 3, 2, 2, 2, 37, 38, 7, 63, 2, 2, 38, 4, 3,
	2, 2, 2, 39, 40, 7, 63, 2, 2, 40, 41, 7, 50, 2, 2, 41, 42, 7, 122, 2, 2,
	42, 6, 3, 2, 2, 2, 43, 44, 7, 101, 2, 2, 44, 45, 7, 110, 2, 2, 45, 46,
	7, 117, 2, 2, 46, 47, 7, 63, 2, 2, 47, 8, 3, 2, 2, 2, 48, 49, 7, 42, 2,
	2, 49, 10, 3, 2, 2, 2, 50, 51, 7, 46, 2, 2, 51, 12, 3, 2, 2, 2, 52, 53,
	7, 43, 2, 2, 53, 14, 3, 2, 2, 2, 54, 56, 9, 2, 2, 2, 55, 54, 3, 2, 2, 2,
	56, 57, 3, 2, 2, 2, 57, 55, 3, 2, 2, 2, 57, 58, 3, 2, 2, 2, 58, 59, 3,
	2, 2, 2, 59, 60, 8, 8, 2, 2, 60, 16, 3, 2, 2, 2, 61, 70, 7, 50, 2, 2, 62,
	66, 9, 3, 2, 2, 63, 65, 9, 4, 2, 2, 64, 63, 3, 2, 2, 2, 65, 68, 3, 2, 2,
	2, 66, 64, 3, 2, 2, 2, 66, 67, 3, 2, 2, 2, 67, 70, 3, 2, 2, 2, 68, 66,
	3, 2, 2, 2, 69, 61, 3, 2, 2, 2, 69, 62, 3, 2, 2, 2, 70, 18, 3, 2, 2, 2,
	71, 73, 9, 5, 2, 2, 72, 71, 3, 2, 2, 2, 73, 74, 3, 2, 2, 2, 74, 72, 3,
	2, 2, 2, 74, 75, 3, 2, 2, 2, 75, 20, 3, 2, 2, 2, 76, 77, 5, 17, 9, 2, 77,
	78, 7, 48, 2, 2, 78, 79, 5, 17, 9, 2, 79, 80, 7, 48, 2, 2, 80, 81, 5, 17,
	9, 2, 81, 82, 7, 48, 2, 2, 82, 83, 5, 17, 9, 2, 83, 84, 7, 49, 2, 2, 84,
	85, 5, 17, 9, 2, 85, 22, 3, 2, 2, 2, 86, 87, 7, 67, 2, 2, 87, 88, 7, 80,
	2, 2, 88, 93, 7, 91, 2, 2, 89, 90, 7, 99, 2, 2, 90, 91, 7, 112, 2, 2, 91,
	93, 7, 123, 2, 2, 92, 86, 3, 2, 2, 2, 92, 89, 3, 2, 2, 2, 93, 24, 3, 2,
	2, 2, 94, 95, 7, 67, 2, 2, 95, 96, 7, 78, 2, 2, 96, 101, 7, 78, 2, 2, 97,
	98, 7, 99, 2, 2, 98, 99, 7, 110, 2, 2, 99, 101, 7, 110, 2, 2, 100, 94,
	3, 2, 2, 2, 100, 97, 3, 2, 2, 2, 101, 26, 3, 2, 2, 2, 102, 103, 7, 80,
	2, 2, 103, 104, 7, 81, 2, 2, 104, 109, 7, 86, 2, 2, 105, 106, 7, 112, 2,
	2, 106, 107, 7, 113, 2, 2, 107, 109, 7, 118, 2, 2, 108, 102, 3, 2, 2, 2,
	108, 105, 3, 2, 2, 2, 109, 28, 3, 2, 2, 2, 110, 111, 7, 85, 2, 2, 111,
	112, 7, 84, 2, 2, 112, 117, 7, 69, 2, 2, 113, 114, 7, 117, 2, 2, 114, 115,
	7, 116, 2, 2, 115, 117, 7, 101, 2, 2, 116, 110, 3, 2, 2, 2, 116, 113, 3,
	2, 2, 2, 117, 30, 3, 2, 2, 2, 118, 119, 7, 70, 2, 2, 119, 120, 7, 85, 2,
	2, 120, 125, 7, 86, 2, 2, 121, 122, 7, 102, 2, 2, 122, 123, 7, 117, 2,
	2, 123, 125, 7, 118, 2, 2, 124, 118, 3, 2, 2, 2, 124, 121, 3, 2, 2, 2,
	125, 32, 3, 2, 2, 2, 126, 127, 7, 70, 2, 2, 127, 128, 7, 85, 2, 2, 128,
	129, 7, 69, 2, 2, 129, 135, 7, 82, 2, 2, 130, 131, 7, 102, 2, 2, 131, 132,
	7, 117, 2, 2, 132, 133, 7, 101, 2, 2, 133, 135, 7, 114, 2, 2, 134, 126,
	3, 2, 2, 2, 134, 130, 3, 2, 2, 2, 135, 34, 3, 2, 2, 2, 136, 137, 7, 86,
	2, 2, 137, 138, 7, 81, 2, 2, 138, 143, 7, 85, 2, 2, 139, 140, 7, 118, 2,
	2, 140, 141, 7, 113, 2, 2, 141, 143, 7, 117, 2, 2, 142, 136, 3, 2, 2, 2,
	142, 139, 3, 2, 2, 2, 143, 36, 3, 2, 2, 2, 15, 2, 57, 66, 69, 72, 74, 92,
	100, 108, 116, 124, 134, 142, 3, 8, 2, 2,
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
	"", "'='", "'=0x'", "'cls='", "'('", "','", "')'",
}

var lexerSymbolicNames = []string{
	"", "", "", "", "", "", "", "WHITESPACE", "DIGITS", "HEX_DIGITS", "NET",
	"ANY", "ALL", "NOT", "SRC", "DST", "DSCP", "TOS",
}

var lexerRuleNames = []string{
	"T__0", "T__1", "T__2", "T__3", "T__4", "T__5", "WHITESPACE", "DIGITS",
	"HEX_DIGITS", "NET", "ANY", "ALL", "NOT", "SRC", "DST", "DSCP", "TOS",
}

type TrafficClassLexer struct {
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

func NewTrafficClassLexer(input antlr.CharStream) *TrafficClassLexer {

	l := new(TrafficClassLexer)

	l.BaseLexer = antlr.NewBaseLexer(input)
	l.Interpreter = antlr.NewLexerATNSimulator(l, lexerAtn, lexerDecisionToDFA, antlr.NewPredictionContextCache())

	l.channelNames = lexerChannelNames
	l.modeNames = lexerModeNames
	l.RuleNames = lexerRuleNames
	l.LiteralNames = lexerLiteralNames
	l.SymbolicNames = lexerSymbolicNames
	l.GrammarFileName = "TrafficClass.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// TrafficClassLexer tokens.
const (
	TrafficClassLexerT__0       = 1
	TrafficClassLexerT__1       = 2
	TrafficClassLexerT__2       = 3
	TrafficClassLexerT__3       = 4
	TrafficClassLexerT__4       = 5
	TrafficClassLexerT__5       = 6
	TrafficClassLexerWHITESPACE = 7
	TrafficClassLexerDIGITS     = 8
	TrafficClassLexerHEX_DIGITS = 9
	TrafficClassLexerNET        = 10
	TrafficClassLexerANY        = 11
	TrafficClassLexerALL        = 12
	TrafficClassLexerNOT        = 13
	TrafficClassLexerSRC        = 14
	TrafficClassLexerDST        = 15
	TrafficClassLexerDSCP       = 16
	TrafficClassLexerTOS        = 17
)