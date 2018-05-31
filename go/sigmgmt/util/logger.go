// Copyright 2018 Anapaya Systems

package util

import (
	"io"

	"github.com/scionproto/scion/go/lib/log"
)

var _ io.Writer = &AccessLogger{}

// AccessLogger logs every access to the standard log file
type AccessLogger struct{}

func (al *AccessLogger) Write(p []byte) (n int, err error) {
	log.Info(string(p))
	return len(p), nil
}
