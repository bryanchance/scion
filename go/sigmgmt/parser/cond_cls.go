// Copyright 2018 Anapaya Systems

package parser

import "github.com/scionproto/scion/go/lib/pktcls"

const typeCondClass = "CondClass"

var _ pktcls.Cond = (*CondClass)(nil)

// CondClass conditions return true if the embedded traffic class returns true
type CondClass struct {
	TrafficClass string
}

func (c CondClass) Eval(v interface{}) bool {
	return false
}

func (c CondClass) Type() string {
	return typeCondClass
}

func NewCondClass(id string) *CondClass {
	return &CondClass{TrafficClass: id}
}
