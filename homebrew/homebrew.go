package homebrew

import (
	"errors"
	"regexp"
	"runtime"
	"strings"

	"github.com/brandmeister/go-brandmeister/dmr"
)

// Debug enables debug messages
var Debug bool

// Errors
var (
	ErrInvalidCallsign       = errors.New("homebrew: invalid callsign")
	ErrInvalidColorCode      = errors.New("homebrew: invalid color code")
	ErrInvalidLatitude       = errors.New("homebrew: invalid latitude")
	ErrInvalidLongitude      = errors.New("homebrew: invalid longitude")
	ErrMasterRefusedLogin    = errors.New("homebrew: master refused login")
	ErrMasterRefusedPassword = errors.New("homebrew: master refused password")
	ErrMasterClose           = errors.New("homebrew: master sent close")
	ErrMasterShortNonce      = errors.New("homebrew: master sent short nonce")
)

var callsignRegex = regexp.MustCompile(`^([A-Z0-9]{0,8})$`)

// Configuration for the Homebrew repeater
type Configuration struct {
	// Callsign of the repeater
	Callsign string

	// ID DMR-ID of the repeater
	ID uint32

	// RXFreq frequency in Hz
	RXFreq int

	// TXFreq frequency in Hz
	TXFreq int

	// TXPower in dBm, decimal [0,99]
	TXPower uint8

	// ColorCode [1,15]
	ColorCode uint8

	// Latitude with north as positive [-90,+90]
	Latitude float64

	// Longitude with east as positive [-180+,180]
	Longitude float64

	// Height above ground in meters
	Height uint16

	// Location description
	Location string

	// Description (optional) about the repeater
	Description string

	// URL (optional) for the repeater or group
	URL string

	// SoftwareID with version (no HTML, no ads, no spam)
	SoftwareID string

	// PackageID with version and platform (no HTML, no ads, no spam)
	PackageID string
}

// Check if the supplied Configuration is sane
func (c *Configuration) Check() error {
	if len(c.Callsign) < 4 || len(c.Callsign) > 8 {
		return ErrInvalidCallsign
	}
	c.Callsign = strings.ToUpper(c.Callsign)
	if !callsignRegex.MatchString(c.Callsign) {
		return ErrInvalidCallsign
	}

	if c.TXPower > 99 {
		c.TXPower = 99
	}

	if c.ColorCode < 0 || c.ColorCode > 15 {
		return ErrInvalidColorCode
	}

	if c.Latitude < -90 || c.Latitude > 90 {
		return ErrInvalidLatitude
	}

	if c.Longitude < -180 || c.Longitude > 180 {
		return ErrInvalidLongitude
	}

	if c.Height > 999 {
		c.Height = 999
	}

	if len(c.Location) > 20 {
		c.Location = c.Location[:20]
	}

	if len(c.Description) > 20 {
		c.Description = c.Description[:20]
	}

	if len(c.URL) > 124 {
		c.URL = c.URL[:124]
	}

	if len(c.SoftwareID) > 40 {
		c.SoftwareID = c.SoftwareID[:40]
	} else if c.SoftwareID == "" {
		c.SoftwareID = "go-brandmeister/homebrew 1.0"
	}

	if len(c.PackageID) > 40 {
		c.PackageID = c.PackageID[:40]
	} else if c.PackageID == "" {
		c.PackageID = runtime.Version()
	}

	return nil
}

func rpad(s string, l int) []byte {
	d := l - len(s)
	if d < 0 {
		return []byte(s[:l])
	}
	if d == 0 {
		return []byte(s)
	}
	p := make([]byte, d)
	for i := range p {
		p[i] = ' '
	}
	return append([]byte(s), p...)
}

type Options uint8

func (o Options) Slot() uint8 {
	return uint8(o) & 0x01
}

func (o Options) Protect() bool {
	return (uint8(o) >> 1) == 1
}

func (o Options) FrameType() uint8 {
	return uint8(o>>2) & 0x03
}

func (o Options) DataType() uint8 {
	return uint8(o >> 4)
}

type DMRData struct {
	Signature [4]byte  // 4
	Sequence  uint8    // 5
	Source    dmr.ID   // 8
	Target    dmr.ID   // 11
	Repeater  uint32   // 15
	Options   Options  // 16
	Stream    uint32   // 20
	Data      [33]byte // 53
}
