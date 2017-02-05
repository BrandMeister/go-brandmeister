package rewind

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"
)

type testServer struct {
	*net.UDPConn
	addr    string
	packets [][]byte
}

func newTestServer(t *testing.T, packets [][]byte) (*testServer, error) {
	s := &testServer{
		packets: packets,
	}

	for i := 0; i < 10; i++ {
		s.addr = fmt.Sprintf(":%d", rand.Intn(65536-1024)+1024)
		a, err := net.ResolveUDPAddr("udp", s.addr)
		if err != nil {
			t.Logf("test server: resolve %s failed: %v\n", s.addr, err)
		}
		if s.UDPConn, err = net.ListenUDP("udp", a); err != nil {
			t.Logf("test server: listen on %s failed: %v\n", s.addr, err)
			s.UDPConn = nil
			continue
		}
	}

	if s.UDPConn == nil {
		return nil, errors.New("failed to find a free UDP port")
	}

	t.Logf("test server: listening on %s\n", s.addr)
	return s, nil
}

func (s *testServer) Run() error {
	b := make([]byte, 128)
	_, peer, err := s.ReadFrom(b)
	if err != nil {
		return err
	}

	for _, packet := range s.packets {
		_, err := s.WriteTo(packet, peer)
		if err != nil {
			return err
		}
	}
	return nil
}

func mustUnhex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func TestConfiguration(t *testing.T) {
	tests := []struct {
		Data []byte
		Test interface{}
	}{
		{
			Data: mustUnhex("524557494e4430310200000000000000"),
			Test: nil,
		},
		{
			Data: []byte{0x52, 0x45, 0x57, 0x49, 0x4e, 0x44, 0x30, 0x31, 0x00, 0x09, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00},
			Test: nil, // internal state, not expecting answer
		},
		{
			Data: mustUnhex("524557494e44303112090100f10200000c000000000000cc1f24b98432d7"),
			Test: nil, // &dmr.TerminatorLC{},
		},
	}

	packets := make([][]byte, len(tests))
	for i, test := range tests {
		packets[i] = test.Data
	}
	s, err := newTestServer(t, packets)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	c, err := NewClient(s.addr, "")
	if err != nil {
		t.Fatal(err)
	}

	p := make(chan Payload)
	go c.ListenAndServe(p)

	if err = s.Run(); err != nil {
		t.Fatal(err)
	}

	for i, test := range tests {
		if test.Test == nil {
			continue
		}
		payload := <-p
		if payload == nil {
			t.Fatalf("test %d failed: %+v", i, test)
		}
	}

	if l := len(p); l != 0 {
		t.Fatalf("%d remains in channel", l)
	}
}

func TestMain(m *testing.M) {
	// For finding random port
	rand.Seed(time.Now().Unix())

	m.Run()
}
