package dmr

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
)

// Frame is a DMR frame
type Frame interface {
}

// ID is a DMR ID
type ID [3]byte

func NewID(id uint32) ID {
	return ID{
		byte(id >> 16),
		byte(id >> 8),
		byte(id),
	}
}

func (id ID) Int() int {
	return int(id.Uint32())
}

func (id ID) String() string {
	return strconv.Itoa(id.Int())
}

func (id ID) Uint32() uint32 {
	return uint32(id[0])<<16 | uint32(id[1])<<8 | uint32(id[2])
}

// LC is a Link Control header
type LC struct {
	Options      uint8
	FeatureSetID uint8
}

// Protect indicates a private call
func (lc LC) Protect() bool {
	return lc.Options&0x80 == 0x80
}

// Opcode returns the FLCO (Full Link Control Opcode)
func (lc LC) Opcode() uint8 {
	return lc.Options & 0x7f
}

// FullLC is a Full Link Control frame
type FullLC struct {
	LC
	ServiceOptions uint8
	Target         ID
	Source         ID
}

// VoiceHeader Link Control
type VoiceHeader struct {
	FullLC
}

// TerminatorLC is a Terminator Link Control frame
type TerminatorLC struct {
	FullLC
}

// EmbeddedData is an Embedded Data frame
type EmbeddedData struct {
	LC
	Target ID
	Source ID
}

// Voice is a voice frame
type Voice [33]byte

// Parse a DMR frame of the given dataType.
func Parse(dataType uint8, data []byte) (Frame, error) {
	var frame interface{}

	switch dataType {
	case TypeVoiceHeader:
		frame = &VoiceHeader{}

	case TypeTerminatorLC:
		frame = &TerminatorLC{}

	case TypeEmbeddedData:
		frame = &EmbeddedData{}

	case TypeVoiceFrameA, TypeVoiceFrameB, TypeVoiceFrameC,
		TypeVoiceFrameD, TypeVoiceFrameE, TypeVoiceFrameF:
		frame = &Voice{}
	}

	if frame == nil {
		return nil, fmt.Errorf("dmr: unable to parse data type %#02x", dataType)
	}

	r := bytes.NewBuffer(data)
	if err := binary.Read(r, binary.BigEndian, frame); err != nil {
		return nil, err
	}

	return frame, nil
}
