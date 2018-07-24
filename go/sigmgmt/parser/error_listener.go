// Copyright 2018 Anapaya Systems

package parser

import (
	"fmt"

	"github.com/antlr/antlr4/runtime/Go/antlr"

	"github.com/scionproto/scion/go/lib/log"
)

type ErrorListener struct {
	*antlr.DefaultErrorListener
	msg       string
	errorType string
}

func (l *ErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line,
	column int, msg string, e antlr.RecognitionException) {
	l.msg = msg
	log.Debug(fmt.Sprintf("%s Error", l.errorType), "err", msg)
}
