// Copyright 2017 Anapaya

// Package aclextn contains the ACL Extension for the Border Router.  It
// initializes a single instance of *ACLMap that can be accessed via Map().
package aclextn

import (
	"sync"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
)

var aclMap = newACLMap()

// Return the ACLMap
func Map() *ACLMap {
	return aclMap
}

// ACLMap maps common.IFIDType to ACL
type ACLMap sync.Map

func newACLMap() *ACLMap {
	return &ACLMap{}
}

func (m *ACLMap) Delete(key common.IFIDType) {
	(*sync.Map)(m).Delete(key)
}

func (m *ACLMap) Load(key common.IFIDType) (ACL, bool) {
	value, ok := (*sync.Map)(m).Load(key)
	if value == nil {
		return nil, ok
	}
	return value.(ACL), ok
}

func (m *ACLMap) LoadOrStore(key common.IFIDType, value ACL) (ACL, bool) {
	actual, ok := (*sync.Map)(m).LoadOrStore(key, value)
	if actual == nil {
		return nil, ok
	}
	return actual.(ACL), ok
}

func (m *ACLMap) Store(key common.IFIDType, value ACL) {
	(*sync.Map)(m).Store(key, value)
}

func (m *ACLMap) Range(f func(key common.IFIDType, value ACL) bool) {
	(*sync.Map)(m).Range(func(key, value interface{}) bool {
		return f(key.(common.IFIDType), value.(ACL))
	})
}

type ACL []addr.IA

// Match returns true if ia is in acl and false otherwise.
func (acl ACL) Match(ia addr.IA) bool {
	for _, aclIA := range acl {
		if ia.Eq(aclIA) {
			return true
		}
	}
	return false
}
