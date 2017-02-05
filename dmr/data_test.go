package dmr

import "testing"

func TestID(t *testing.T) {
	tests := map[uint32]string{
		204:      "204",
		16777215: "16777215",
	}

	for i, s := range tests {
		id := NewID(i)
		if id.String() != s {
			t.Fail()
		}
	}
}

func TestParseVoiceHeader(t *testing.T) {
	tests := []struct {
		Data []byte
		Test VoiceHeader
	}{
		{
			Data: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0xcc, 0x1f, 0x24, 0xb9, 0x8b, 0x3d, 0xd8},
			Test: VoiceHeader{FullLC{Target: NewID(204), Source: NewID(2041017)}},
		},
	}

	for _, test := range tests {
		f, err := Parse(TypeVoiceHeader, test.Data)
		if err != nil {
			t.Fatal(err)
		}
		h, ok := f.(*VoiceHeader)
		if !ok {
			t.Fail()
		}
		if h.Options != test.Test.Options {
			t.Fail()
		}
		if h.Protect() != test.Test.Protect() {
			t.Fail()
		}
		if h.Opcode() != test.Test.Opcode() {
			t.Fail()
		}
		if h.ServiceOptions != test.Test.ServiceOptions {
			t.Fail()
		}
		if h.FeatureSetID != test.Test.FeatureSetID {
			t.Fail()
		}
		if h.Target.Int() != test.Test.Target.Int() {
			t.Fail()
		}
		if h.Source.Int() != test.Test.Source.Int() {
			t.Fail()
		}
	}
}

func TestParseTerminatorLC(t *testing.T) {
	tests := []struct {
		Data []byte
		Test TerminatorLC
	}{
		{
			Data: []byte{0x00, 0x00, 0x00, 0x00, 0x01, 0x36, 0x30, 0x07, 0x17, 0x70, 0x94, 0x07},
			Test: TerminatorLC{FullLC{Target: NewID(310), Source: NewID(3147543)}}},
	}

	for _, test := range tests {
		f, err := Parse(TypeTerminatorLC, test.Data)
		if err != nil {
			t.Fatal(err)
		}
		h, ok := f.(*TerminatorLC)
		if !ok {
			t.Fail()
		}
		if h.Options != test.Test.Options {
			t.Fail()
		}
		if h.Protect() != test.Test.Protect() {
			t.Fail()
		}
		if h.Opcode() != test.Test.Opcode() {
			t.Fail()
		}
		if h.ServiceOptions != test.Test.ServiceOptions {
			t.Fail()
		}
		if h.FeatureSetID != test.Test.FeatureSetID {
			t.Fail()
		}
		if h.Target.Int() != test.Test.Target.Int() {
			t.Fail()
		}
		if h.Source.Int() != test.Test.Source.Int() {
			t.Fail()
		}
	}
}

func TestParseError(t *testing.T) {
	t.Run("invalid dataType", func(t *testing.T) {
		f, err := Parse(0xff, []byte{0xff})
		if err == nil {
			t.Fail()
		}
		if f != nil {
			t.Fail()
		}
	})

	t.Run("empty buffer", func(t *testing.T) {
		f, err := Parse(TypeEmbeddedData, nil)
		if err == nil {
			t.Fail()
		}
		if f != nil {
			t.Fail()
		}
	})
}
