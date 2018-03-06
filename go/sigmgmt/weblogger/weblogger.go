// Copyright 2018 Anapaya Systems

// Package weblogger contains an html-friendly renderer for log15 messages.
package weblogger

import (
	"fmt"
	"strings"

	log "github.com/inconshreveable/log15"
)

// handler is a buffered logging handler that outputs an html-friendly of the
// log on calls to Flush.
type handler struct {
	Items []*log.Record
}

func (l *handler) Log(r *log.Record) error {
	l.Items = append(l.Items, r)
	return nil
}

type Logger struct {
	log.Logger
	handler *handler
}

// New returns a new web logger. Messages logged to the new logger will also be
// logged to the parent. Call Records in templates to gain access to the
// records and render the logs.
func New(parent log.Logger) *Logger {
	child := parent.New()
	newHandler := &handler{}
	child.SetHandler(
		log.MultiHandler(
			parent.GetHandler(),
			newHandler,
		),
	)
	return &Logger{
		Logger:  child,
		handler: newHandler,
	}
}

func (l *Logger) Records() []*log.Record {
	return l.handler.Items
}

// pairs constructs a slice of key-value pairs out of a log15 context slice. If
// odd-length, the last key is ignored.
func Pairs(ctxs []interface{}) []kvPair {
	if len(ctxs)%2 == 1 {
		ctxs = ctxs[:len(ctxs)-1]
	}
	var pairs []kvPair
	for i := 0; i < len(ctxs); i += 2 {
		pairs = append(pairs, kvPair{Key: ctxs[i], Value: ctxs[i+1]})
	}
	return pairs
}

type kvPair struct {
	Key   interface{}
	Value interface{}
}

func PrintLogLevel(level log.Lvl) string {
	return strings.ToUpper(fmt.Sprintf("%v", level))
}
