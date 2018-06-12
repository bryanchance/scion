// Copyright 2018 Anapaya Systems

package parser

import (
	"github.com/antlr/antlr4/runtime/Go/antlr"
)

type ErrorListener struct {
	*antlr.DefaultErrorListener
	msg string
}

func (c *ErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line,
	column int, msg string, e antlr.RecognitionException) {
	c.msg = msg
}
