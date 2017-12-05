// Copyright 2017 Anapaya

package extn

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

type Extension interface {
	Pack() (common.RawBytes, error)
	CtrlExtnType() common.RawBytes
	proto.Cerealizable
}

var _ proto.Cerealizable = (*CtrlExtnDataList)(nil)

type CtrlExtnDataList struct {
	Items []*CtrlExtnData
}

func NewCtrlExtnDataListFromValues(items []*CtrlExtnData) *CtrlExtnDataList {
	return &CtrlExtnDataList{Items: items}
}

func NewCtrlExtnDataListFromRaw(b common.RawBytes) (*CtrlExtnDataList, error) {
	edl := &CtrlExtnDataList{}
	return edl, proto.ParseFromRaw(edl, edl.ProtoId(), b)
}

func (edl *CtrlExtnDataList) ProtoId() proto.ProtoIdType {
	return proto.CtrlExtnDataList_TypeID
}

func (edl *CtrlExtnDataList) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(edl, b)
}

func (edl *CtrlExtnDataList) String() string {
	return fmt.Sprintf("Items: %v", edl.Items)
}

var _ proto.Cerealizable = (*CtrlExtnData)(nil)

type CtrlExtnData struct {
	Type common.RawBytes
	Data common.RawBytes
}

func NewCtrlExtnDataFromValues(e Extension, arenaSize int) (*CtrlExtnData, error) {
	raw, err := e.Pack()
	if err != nil {
		return nil, common.NewCError("Unable to pack extension", "extn", e, "err", err)
	}
	return &CtrlExtnData{Type: e.CtrlExtnType(), Data: raw}, nil
}

func NewCtrlExtnDataFromRaw(b common.RawBytes) (*CtrlExtnData, error) {
	ed := &CtrlExtnData{}
	return ed, proto.ParseFromRaw(ed, ed.ProtoId(), b)
}

func (ed *CtrlExtnData) ProtoId() proto.ProtoIdType {
	return proto.CtrlExtnData_TypeID
}

func (ed *CtrlExtnData) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(ed, b)
}

func (ed *CtrlExtnData) TypeStr() string {
	return string(ed.Type)
}

func (ed *CtrlExtnData) String() string {
	return fmt.Sprintf("Type: %v, Data length: %v", ed.Type, len(ed.Data))
}
