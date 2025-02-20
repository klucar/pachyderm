// Code generated by protoc-gen-go.
// source: testing/testing.proto
// DO NOT EDIT!

package protolog_testing

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// discarding unused import protolog "go.pedge.io/protolog"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Foo struct {
	StringField string `protobuf:"bytes,1,opt,name=string_field" json:"string_field,omitempty"`
	Int32Field  int32  `protobuf:"varint,2,opt,name=int32_field" json:"int32_field,omitempty"`
}

func (m *Foo) Reset()         { *m = Foo{} }
func (m *Foo) String() string { return proto.CompactTextString(m) }
func (*Foo) ProtoMessage()    {}

type Bar struct {
	StringField string `protobuf:"bytes,1,opt,name=string_field" json:"string_field,omitempty"`
	Int32Field  int32  `protobuf:"varint,2,opt,name=int32_field" json:"int32_field,omitempty"`
}

func (m *Bar) Reset()         { *m = Bar{} }
func (m *Bar) String() string { return proto.CompactTextString(m) }
func (*Bar) ProtoMessage()    {}

type Baz struct {
	Bat *Baz_Bat `protobuf:"bytes,1,opt,name=bat" json:"bat,omitempty"`
}

func (m *Baz) Reset()         { *m = Baz{} }
func (m *Baz) String() string { return proto.CompactTextString(m) }
func (*Baz) ProtoMessage()    {}

func (m *Baz) GetBat() *Baz_Bat {
	if m != nil {
		return m.Bat
	}
	return nil
}

type Baz_Bat struct {
	Ban *Baz_Bat_Ban `protobuf:"bytes,1,opt,name=ban" json:"ban,omitempty"`
}

func (m *Baz_Bat) Reset()         { *m = Baz_Bat{} }
func (m *Baz_Bat) String() string { return proto.CompactTextString(m) }
func (*Baz_Bat) ProtoMessage()    {}

func (m *Baz_Bat) GetBan() *Baz_Bat_Ban {
	if m != nil {
		return m.Ban
	}
	return nil
}

type Baz_Bat_Ban struct {
	StringField string `protobuf:"bytes,1,opt,name=string_field" json:"string_field,omitempty"`
	Int32Field  int32  `protobuf:"varint,2,opt,name=int32_field" json:"int32_field,omitempty"`
}

func (m *Baz_Bat_Ban) Reset()         { *m = Baz_Bat_Ban{} }
func (m *Baz_Bat_Ban) String() string { return proto.CompactTextString(m) }
func (*Baz_Bat_Ban) ProtoMessage()    {}

type Empty struct {
}

func (m *Empty) Reset()         { *m = Empty{} }
func (m *Empty) String() string { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()    {}
