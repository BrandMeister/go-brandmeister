package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"

	"github.com/brandmeister/go-brandmeister/dmr"
	"github.com/brandmeister/go-brandmeister/rewind"
)

var dmrDataType = map[uint8]string{
	dmr.TypePIHeader:        "pi header",        // 0x00
	dmr.TypeVoiceHeader:     "voice header",     // 0x01
	dmr.TypeTerminatorLC:    "terminator lc",    // 0x02
	dmr.TypeCSBK:            "csbk",             // 0x03
	dmr.TypeMBCHeader:       "mbc header",       // 0x04
	dmr.TypeMBCContinuation: "mbc continuation", // 0x05
	dmr.TypeDataHeader:      "data header",      // 0x06
	dmr.TypeRate12Data:      "rate 1/2 data",    // 0x07
	dmr.TypeRate34Data:      "rate 3/4 data",    // 0x08
	dmr.TypeIdle:            "idle",             // 0x09
	dmr.TypeEmbeddedData:    "embedded data",    // 0x11
}

func main() {
	id := flag.Int("id", 1234, "application id")
	addr := flag.String("addr", "brandmeister.pd0zry.ampr.org", "server address")
	pass := flag.String("password", "", "server password")
	subs := flag.String("subscribe", "", "subscribe to TGs")
	debug := flag.Bool("debug", false, "enable protocol debugging")
	flag.Parse()

	rewind.Debug = *debug

	c, err := rewind.NewClient(*addr, *pass)
	if err != nil {
		log.Fatalln(err)
	}

	c.Description = fmt.Sprintf("go-brandmeister rewind-dump.go (%s; %s; %s; %s)",
		runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.Compiler)
	c.RemoteID = uint32(*id)
	c.Options = rewind.OptionSuperHeader

	c.Subscriptions = make(map[uint32]rewind.SessionType)
	for _, s := range strings.Split(*subs, ",") {
		s = strings.TrimSpace(s)
		if tg, err := strconv.ParseUint(s, 10, 32); err != nil {
			log.Fatalf("error parsing %q: %v\n", s, err)
		} else if tg > 0 {
			c.Subscriptions[uint32(tg)] = rewind.GroupVoice
		}
	}

	p := make(chan rewind.Payload)
	go func(p <-chan rewind.Payload) {
		for payload := range p {
			switch payload := payload.(type) {
			case *rewind.DMRData:
				t, ok := dmrDataType[payload.Type]
				if !ok {
					t = fmt.Sprintf("unknown %#02x", payload.Type)
				}
				f, err := dmr.Parse(payload.Type, payload.Data)
				if err != nil {
					log.Printf("received: DMR %s (%v)\n", t, err)
				} else {
					log.Printf("received: DMR %s: %T %v\n", t, f, f)
				}

			case *rewind.DMRAudio:
				log.Println("received: DMR audio frame")

			case *rewind.SuperHeader:
				log.Printf("received: super header: %T %v\n", payload, payload)

			case *rewind.Raw:
				log.Printf("received: raw (unparsed) type %d\n", payload.Type)
			}
		}
	}(p)

	log.Fatalln(c.ListenAndServe(p))
}
