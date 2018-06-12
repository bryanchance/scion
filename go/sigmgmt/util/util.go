// Copyright 2018 Anapaya Systems

package util

import (
	"regexp"

	"github.com/scionproto/scion/go/lib/common"
)

const (
	nameErrorMsg    = "bad string, must only contain alphanumeric characters, '.', '_', and '-'"
	userValErrorMsg = "bad user, must only contain alphanumeric characters"
	keyValErrorMsg  = "bad key, must only contain alphanumeric characters, '.', '/', '_', and '-'"
)

var (
	validateHostRe *regexp.Regexp
	validateUserRe *regexp.Regexp
	validateKeyRe  *regexp.Regexp
)

func init() {
	validateHostRe = regexp.MustCompile("^[a-zA-Z0-9-_.]+$")
	validateUserRe = regexp.MustCompile("^[a-zA-Z0-9]+$")
	validateKeyRe = regexp.MustCompile("^[a-zA-Z0-9-_./]+$")
}

// ValidateIdentifier returns an error if str is not of format [a-zA-Z0-9.-_]+.
func ValidateIdentifier(str string) error {
	matched := validateHostRe.MatchString(str)
	if !matched {
		return common.NewBasicError(nameErrorMsg, nil, "token", str)
	}
	return nil
}

// ValidateUser returns an error if str is not of format [a-zA-Z0-9]+.
func ValidateUser(str string) error {
	matched := validateUserRe.MatchString(str)
	if !matched {
		return common.NewBasicError(userValErrorMsg, nil, "token", str)
	}
	return nil
}

// ValidateKey returns an error if str is not of format [a-zA-Z0-9.-_/]+.
func ValidateKey(str string) error {
	matched := validateKeyRe.MatchString(str)
	if !matched {
		return common.NewBasicError(keyValErrorMsg, nil, "token", str)
	}
	return nil
}
