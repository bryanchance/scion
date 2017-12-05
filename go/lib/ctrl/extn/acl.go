// Copyright 2017 Anapaya

package extn

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

var _ Extension = (*PushACL)(nil)

type PushACL struct {
	Permit []addr.IAInt `capnp:"permit"`
}

func NewPushACLFromValues(acl []*addr.ISD_AS) *PushACL {
	msg := &PushACL{}
	for i := range acl {
		msg.Permit = append(msg.Permit, acl[i].IAInt())
	}
	return msg
}

func NewPushACLFromRaw(b common.RawBytes) (*PushACL, error) {
	acl := &PushACL{}
	return acl, proto.ParseFromRaw(acl, acl.ProtoId(), b)
}

func (acl *PushACL) ProtoId() proto.ProtoIdType {
	return proto.PushACL_TypeID
}

func (acl *PushACL) CtrlExtnType() common.RawBytes {
	return common.RawBytes("com.anapaya.pushacl")
}

func (acl *PushACL) Pack() (common.RawBytes, error) {
	return proto.PackRoot(acl)
}

func (acl *PushACL) String() string {
	return fmt.Sprintf("Permit: %v", acl.Permit)
}

func (acl *PushACL) ACL() []*addr.ISD_AS {
	v := make([]*addr.ISD_AS, len(acl.Permit))
	for i := range acl.Permit {
		v[i] = addr.IAInt(acl.Permit[i]).IA()
	}
	return v
}
