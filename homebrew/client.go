package homebrew

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

type authState uint8

const (
	authNone authState = iota
	authSentLogin
	authSentKey
	authDone
	authFailed
)

// Client implementing the Homebrew protocol
type Client struct {
	KeepAlive time.Duration
	Timeout   time.Duration

	// Configuration of the repeater
	Configuration *Configuration

	// Description of the client
	Description string

	// URL of the client
	URL string

	conn net.Conn
	data chan []byte
	quit chan struct{}
	errs chan error

	password    string
	auth        authState
	hexid       [8]byte
	nonce       [8]byte
	pingSent    time.Time
	pingLatency time.Duration
}

// NewClient sets up a Homebrew protocol client with defaults configured.
func NewClient(cfg *Configuration, addr, password string) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("homebrew: nil configuration")
	}
	if err := cfg.Check(); err != nil {
		return nil, err
	}

	c := &Client{
		KeepAlive:     DefaultKeepAliveInterval,
		Timeout:       DefaultTimeout,
		Configuration: cfg,
		password:      password,
	}

	if !strings.Contains(addr, ":") {
		addr = fmt.Sprintf("%s:%d", addr, DefaultPort)
	}

	if Debug {
		log.Printf("homebrew: connecting to udp://%s\n", addr)
	}

	var err error
	if c.conn, err = net.Dial("udp", addr); err != nil {
		return nil, err
	}

	return c, nil
}

// Close the client socket and stop the receiver loop after it has been started
// by ListenAndServe.
func (c *Client) Close() error {
	if c.quit != nil {
		c.quit <- struct{}{} // listen
		c.quit <- struct{}{} // receive
	}
	return c.conn.Close()
}

// ListenAndServe starts the packet receiver and decoder
func (c *Client) ListenAndServe(f chan<- *DMRData) error {
	c.quit = make(chan struct{}, 2)
	c.data = make(chan []byte)
	c.errs = make(chan error)

	go c.receive()

	if err := c.sendLogin(); err != nil {
		return err
	}

	var (
		timeout = time.NewTicker(c.Timeout)
		last    = time.Now()
	)

	for {
		if c.auth == authDone && time.Since(c.pingSent) > c.KeepAlive {
			if err := c.sendPing(); err != nil {
				return err
			}
		}

		select {
		case data := <-c.data:
			if err := c.parse(data, f); err != nil {
				c.quit <- struct{}{} // signal receiver
				return err
			}

			last = time.Now()
			timeout.Stop()
			timeout = time.NewTicker(c.Timeout)

		case <-time.After(c.KeepAlive):
			if c.auth == authDone {
				if err := c.sendPing(); err != nil {
					c.quit <- struct{}{} // signal receiver
					return err
				}
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

func (c *Client) WriteDMR(dmrData *DMRData) error {
	copy(dmrData.Signature[:], []byte(SignDMRData))
	return binary.Write(c.conn, binary.BigEndian, dmrData)
}

func (c *Client) parse(b []byte, f chan<- *DMRData) (err error) {
	if len(b) < 3 {
		return io.ErrShortBuffer
	}

	switch c.auth {
	case authSentLogin:
		// We expect MSTACK or MSTNAK
		if bytes.Equal(b[:len(SignMasterNAK)], []byte(SignMasterNAK)) {
			c.auth = authFailed
			return ErrMasterRefusedLogin
		} else if bytes.Equal(b[:len(SignMasterACK)], []byte(SignMasterACK)) {
			if n := copy(c.nonce[:], b[len(SignMasterACK)+8:]); n != 8 {
				c.auth = authFailed
				return ErrMasterShortNonce
			}
			if Debug {
				log.Println("homebrew: received nonce, sending password")
			}
			return c.sendKey()
		}

		// Ignored
		log.Printf("homebrew: %q\n", b)
		return nil

	case authSentKey:
		// We expect MSTACK or MSTNAK
		if bytes.Equal(b[:len(SignMasterNAK)], []byte(SignMasterNAK)) {
			c.auth = authFailed
			return ErrMasterRefusedPassword
		} else if bytes.Equal(b[:len(SignMasterACK)], []byte(SignMasterACK)) {
			if Debug {
				log.Println("homebrew: logged in, sending configuration")
			}
			c.auth = authDone
			return c.sendConfiguration()
		}

		// Ignored
		log.Printf("homebrew: %q\n", b)
		return nil
	}

	switch {
	case bytes.Equal(b[:len(SignDMRData)], []byte(SignDMRData)):
		var dmrData DMRData
		if derr := binary.Read(bytes.NewBuffer(b), binary.BigEndian, &dmrData); derr != nil {
			log.Printf("homebrew: failed to decode %d bytes of DMRD: %v\n", len(b), derr)
			return
		}
		f <- &dmrData
		return nil

	case bytes.Equal(b[:len(SignMasterClose)], []byte(SignMasterClose)):
		return ErrMasterClose

	case bytes.Equal(b[:len(SignMasterACK)], []byte(SignMasterACK)):
		if Debug {
			log.Println("homebrew: configuration accepted by master")
		}
		return c.sendPing()

	case bytes.Equal(b[:len(SignMasterNAK)], []byte(SignMasterNAK)):
		log.Println("homebrew: master dropped connection, logging in")
		return c.sendLogin()

	case bytes.Equal(b[:len(SignRepeaterPong)], []byte(SignRepeaterPong)):
		c.pingLatency = time.Since(c.pingSent)
		if Debug {
			log.Printf("homebrew: ping RTT %s\n", c.pingLatency)
		}
		return nil
	}

	log.Printf("homebrew: %q\n", b)

	return nil
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

func (c *Client) sendLogin() error {
	copy(c.hexid[:], []byte(fmt.Sprintf("%08x", c.Configuration.ID)))
	var (
		data = make([]byte, len(SignRepeaterLogin)+8)
		n    = copy(data, SignRepeaterLogin)
	)
	copy(data[n:], c.hexid[:])
	c.auth = authSentLogin
	_, err := c.conn.Write(data)
	return err
}

func (c *Client) sendKey() error {
	var (
		hash = sha256.Sum256(append(c.nonce[:], []byte(c.password)...))
		data = make([]byte, len(SignRepeaterKey)+72)
		n    = copy(data, SignRepeaterKey)
	)
	n += copy(data[n:], c.hexid[:])
	copy(data[n:], []byte(hex.EncodeToString(hash[:])))
	c.auth = authSentKey
	_, err := c.conn.Write(data)
	return err
}

func (c *Client) sendConfiguration() error {
	if err := c.Configuration.Check(); err != nil {
		return err
	}
	var data []byte
	data = []byte(SignRepeaterConfig)
	data = append(data, []byte(fmt.Sprintf("%-8s", c.Configuration.Callsign))...)
	data = append(data, []byte(fmt.Sprintf("%08x", c.Configuration.ID))...)
	data = append(data, []byte(fmt.Sprintf("%09d", c.Configuration.RXFreq))...)
	data = append(data, []byte(fmt.Sprintf("%09d", c.Configuration.TXFreq))...)
	data = append(data, []byte(fmt.Sprintf("%02d", c.Configuration.TXPower))...)
	data = append(data, []byte(fmt.Sprintf("%02d", c.Configuration.ColorCode))...)
	data = append(data, []byte(fmt.Sprintf("%-08f", c.Configuration.Latitude)[:8])...)
	data = append(data, []byte(fmt.Sprintf("%-09f", c.Configuration.Longitude)[:9])...)
	data = append(data, []byte(fmt.Sprintf("%03d", c.Configuration.Height))...)
	data = append(data, []byte(fmt.Sprintf("%-20s", c.Configuration.Location))...)
	data = append(data, []byte(fmt.Sprintf("%-20s", c.Configuration.Description))...)
	data = append(data, []byte(fmt.Sprintf("%-124s", c.Configuration.URL))...)
	data = append(data, []byte(fmt.Sprintf("%-40s", c.Configuration.SoftwareID))...)
	data = append(data, []byte(fmt.Sprintf("%-40s", c.Configuration.PackageID))...)
	if Debug {
		log.Printf("homebrew: sending %+v\n", c.Configuration)
	}
	_, err := c.conn.Write(data)
	return err
}

func (c *Client) sendPing() error {
	var (
		data = make([]byte, len(SignMasterPing)+8)
		n    = copy(data, SignMasterPing)
	)
	copy(data[n:], c.hexid[:])
	_, err := c.conn.Write(data)
	c.pingSent = time.Now()
	return err
}
