// Copyright 2018 Anapaya Systems

package util

import (
	"regexp"

	"github.com/scionproto/scion/go/lib/common"
)

const (
	validationErrorMsg = "bad host name, must only contain alphanumeric characters, " +
		"'.', '_', and '-'"
)

var (
	validateIdentifierRe *regexp.Regexp
)

func init() {
	validateIdentifierRe = regexp.MustCompile("^[a-zA-Z0-9-_.]+$")
}

// ValidateIdentifier returns an error if str is not of format [a-zA-Z0-9.-_]+.
func ValidateIdentifier(str string) error {
	matched := validateIdentifierRe.MatchString(str)
	if !matched {
		return common.NewBasicError(validationErrorMsg, nil, "token", str)
	}
	return nil
}
