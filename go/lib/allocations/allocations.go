// Copyright 2018 Anapaya Systems

// Package allocations contains messages used by IPScraper and IPProvider.
package allocations

import (
	"fmt"

	"github.com/scionproto/scion/go/proto"
)

var _ proto.Cerealizable = (*Request)(nil)

type Allocation struct {
	IA       string `capnp:"ia"`
	Networks []string
}

type Request struct {
	ID string `capnp:"id"`
}

func NewRequest(id string) *Request {
	return &Request{ID: id}
}

func (p *Request) ProtoId() proto.ProtoIdType {
	return proto.AllocationsRequest_TypeID
}

func (p *Request) String() string {
	return fmt.Sprintf("ID: %s", p.ID)
}

var _ proto.Cerealizable = (*Reply)(nil)

type Reply struct {
	ID          string `capnp:"id"`
	Allocations []*Allocation
}

func NewReply(id string, allocations []*Allocation) *Reply {
	return &Reply{ID: id, Allocations: allocations}
}

func (p *Reply) ProtoId() proto.ProtoIdType {
	return proto.AllocationsReply_TypeID
}

func (p *Reply) String() string {
	return fmt.Sprintf("ID: %s Allocations: %s", p.ID, p.Allocations)
}
