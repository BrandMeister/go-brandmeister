package rewind

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strings"
	"sync/atomic"
	"time"
)

type Client struct {
	KeepAlive time.Duration
	Timeout   time.Duration

	// Description of the client
	Description string

	// RemoteID of the client
	RemoteID uint32

	// Options
	Options uint32

	// Subscriptions is a map of Target ID and session type
	Subscriptions map[uint32]SessionType

	// ApplicationCallback for application data
	ApplicationCallback func(dataType uint16, parsed interface{})

	// DeviceCallback for device data
	DeviceCallback func(dataType uint16, parsed interface{})

	conn net.Conn

	quit chan struct{}
	data chan []byte
	errs chan error

	password string
	sequence uint32
	auth     bool
}

// NewClient sets up a Rewind protocol client with defaults
// configured.
func NewClient(addr, password string) (*Client, error) {
	c := &Client{
		KeepAlive: DefaultKeepAliveInterval,
		Timeout:   DefaultTimeout,
		password:  password,
	}

	if !strings.Contains(addr, ":") {
		addr = fmt.Sprintf("%s:%d", addr, DefaultPort)
	}

	if Debug {
		log.Printf("rewind: connecting to udp://%s\n", addr)
	}

	var err error
	if c.conn, err = net.Dial("udp", addr); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) Close() error {
	if c.quit != nil {
		c.quit <- struct{}{} // listen
		c.quit <- struct{}{} // receive
	}
	return c.conn.Close()
}

func (c *Client) ListenAndServe() error {
	c.quit = make(chan struct{}, 2)
	c.data = make(chan []byte)
	c.errs = make(chan error)

	go c.receive()

	if err := c.sendKeepAlive(); err != nil {
		return err
	}

	var (
		timeout = time.NewTicker(c.Timeout)
		last    = time.Now()
	)

	for {
		select {
		case data := <-c.data:
			if len(data) < SignLength {
				if Debug {
					log.Printf("rewind: packet of size %d too short (ignored)\n", len(data))
				}
				continue
			}
			if !bytes.Equal(data[:SignLength], []byte(Sign)) {
				if Debug {
					log.Printf("rewind: %q does not match sign %q (ignored)\n", data[:SignLength], Sign)
				}
				continue
			}
			if err := c.parse(data); err != nil {
				c.quit <- struct{}{} // signal receiver
				return err
			}

			last = time.Now()
			timeout.Stop()
			timeout = time.NewTicker(c.Timeout)

		case <-time.After(c.KeepAlive):
			if err := c.sendKeepAlive(); err != nil {
				c.quit <- struct{}{} // signal receiver
				return err
			}

		case <-timeout.C:
			c.quit <- struct{}{} // signal receiver
			return fmt.Errorf("timeout after %s", time.Since(last))

		case <-c.quit:
			return nil

		case err := <-c.errs:
			return err
		}
	}
}

func (c *Client) receive() {
	for {
		data := make([]byte, 128)
		n, err := c.conn.Read(data)
		if err != nil {
			c.errs <- err
			return
		}
		c.data <- data[:n]
	}
}

func (c *Client) parse(b []byte) (err error) {
	var header Data
	if err = binary.Read(bytes.NewBuffer(b), binary.LittleEndian, &header); err != nil {
		return
	}

	switch header.Type {
	case TypeClose:
		log.Println("rewind: server sent close")
		c.auth = false

	case TypeKeepAlive:
		if !c.auth {
			return c.sendConfiguration()
		}

	case TypeConfiguration:
		if Debug {
			log.Println("rewind: configuration accepted")
		}

		// We're authenticated
		c.auth = true

		for id, typ := range c.Subscriptions {
			if err := c.Subscribe(id, typ); err != nil {
				return err
			}
		}

	case TypeSubscription:
		if Debug {
			log.Println("rewind: subscription confirmed")
		}

	case TypeReport:
		if Debug {
			log.Printf("rewind: received report %q\n", b[DataLength:])
		}

	case TypeChallenge:
		c.auth = false
		return c.sendChallengeResponse(b[DataLength:])

		// Application Data

	case TypeSuperHeader: // 0x0928
		if c.ApplicationCallback == nil {
			return nil
		}

		if l := len(b[DataLength:]); l < SuperHeaderLength {
			log.Printf("rewind: received corrupt super header with length %d (expected %d)\n", l, SuperHeaderLength)
			return nil
		}

		var superHeader SuperHeader
		if derr := binary.Read(bytes.NewBuffer(b[DataLength:]), binary.LittleEndian, &superHeader); derr != nil {
			log.Printf("rewind: unable to unmarshal super header: %v\n", derr)
			return nil
		}

		c.ApplicationCallback(header.Type, &superHeader)

	default:
		if header.Type >= ClassApplication {
			// Application data will be sent to our callback
			if c.ApplicationCallback != nil {
				c.ApplicationCallback(header.Type, b[DataLength:])
			}
		} else if header.Type >= ClassDeviceData {
			// Device data will be sent to our callback
			if c.DeviceCallback != nil {
				c.DeviceCallback(header.Type, b[DataLength:])
			}
		} else if Debug {
			log.Printf("rewind: received unknown packet type %#04x\n", header.Type)
		}

	}

	return
}

// Subscribe to a destination DMR ID and Session type.
func (c *Client) Subscribe(id uint32, typ SessionType) error {
	return c.sendData(TypeSubscription, &SubscriptionData{
		Type:   typ,
		Target: id,
	})
}

func (c *Client) sendData(typ uint16, payload Payload) error {
	header := Data{
		Type:     typ,
		Sequence: atomic.AddUint32(&c.sequence, 1),
		Length:   uint16(payload.Len()),
	}
	copy(header.Sign[:], Sign)

	buffer := new(bytes.Buffer)
	if err := binary.Write(buffer, binary.LittleEndian, header); err != nil {
		return err
	}
	if err := binary.Write(buffer, binary.LittleEndian, payload); err != nil {
		return err
	}

	_, err := c.conn.Write(buffer.Bytes())
	return err
}

func (c *Client) sendChallengeResponse(challenge []byte) error {
	digest := sha256.Sum256(append(challenge, []byte(c.password)...))
	return c.sendData(TypeAuthentication, payload(digest[:]))
}

func (c *Client) sendConfiguration() error {
	log.Printf("rewind: sending configuration (%d)\n", c.Options)
	return c.sendData(TypeConfiguration, ConfigurationData(c.Options))
}

func (c *Client) sendKeepAlive() error {
	var description [DescriptionLength]byte
	copy(description[:], c.Description)
	return c.sendData(TypeKeepAlive, &VersionData{
		RemoteID:    c.RemoteID,
		Service:     ServiceSimpleApplication,
		Description: description,
	})
}
