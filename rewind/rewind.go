// Package rewind implements the Simple External Application rewind protocol as
// described at https://wiki.brandmeister.network/index.php/Simple_External_Application
package rewind

// Debug enables debug messages
var Debug bool

// Packet sizes
const (
	DataLength        = SignLength + 10
	DescriptionLength = 96
)

// Payload to be sent along with data
type Payload interface {
	Len() int
}

type payload []byte

func (p payload) Len() int {
	return len(p)
}

// Data header
type Data struct {
	// Sign of protocol
	Sign [SignLength]byte

	// Type of data
	Type uint16

	// Flags for data
	Flags uint16

	// Sequence number
	Sequence uint32

	// Length of data
	Length uint16
}

// VersionData informs the server about the client
type VersionData struct {
	// RemoteID is the remote application ID
	RemoteID uint32

	// Service type
	Service uint8

	// Description of software and version
	Description [DescriptionLength]byte
}

// Len is the packet size
func (vd VersionData) Len() int {
	var i int
	for ; i < DescriptionLength; i++ {
		if vd.Description[i] != 0 {
			break
		}
	}
	return 5 + i
}

// Generic Data Structures

// Address is a struct in_addr / in_addr_t
type Address uint32

// AddressData contains a network address
type AddressData struct {
	Address Address
	Port    uint16
}

type BindingData []uint16

// Simple Application Protocol

// ConfigurationData contains Options
type ConfigurationData uint32

func (c ConfigurationData) Len() int {
	return 4
}

// SessionType type of transmission
type SessionType uint32

// SubscriptionData contains Session types
type SubscriptionData struct {
	// Type of session
	Type SessionType

	// Target ID
	Target uint32
}

func (d SubscriptionData) Len() int {
	return 8
}

// SuperHeader contains metadata about a transmission
type SuperHeader struct {
	// Type of session
	Type uint32

	// Source ID
	Source uint32

	// Target ID
	Target uint32

	// SourceCall is the source call (or zeros)
	SourceCall [CallLength]byte

	// TargetCall is the target call (or zeros)
	TargetCall [CallLength]byte
}

// Rewind Transport Layer
