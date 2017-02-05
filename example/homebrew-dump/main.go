package main

import (
	"flag"
	"log"

	"github.com/brandmeister/go-brandmeister/dmr"
	"github.com/brandmeister/go-brandmeister/homebrew"
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

func handle(f <-chan *homebrew.DMRData) {
	for data := range f {
		switch data.Options.FrameType() {
		case 0x00:
			log.Printf("dmr: %s->%s voice frame %c\n", data.Source, data.Target, 'A'+data.Options.DataType())
		case 0x01:
			log.Printf("dmr: %s->%s voice sync\n", data.Source, data.Target)
		case 0x02:
			log.Printf("dmr: %s->%s %s\n", data.Source, data.Target, dmrDataType[data.Options.DataType()])
		}
	}
}

func main() {
	id := flag.Int("id", 1234, "DMR ID")
	call := flag.String("call", "N0CALL", "callsign")
	addr := flag.String("addr", "brandmeister.pd0zry.ampr.org", "server address")
	pass := flag.String("password", "", "server password")
	debug := flag.Bool("debug", false, "enable protocol debugging")
	flag.Parse()

	homebrew.Debug = *debug

	config := &homebrew.Configuration{
		Callsign:    *call,
		ID:          uint32(*id),
		SoftwareID:  "go-brandmeister/example/homebrew-dump",
		RXFreq:      430000000,
		TXFreq:      430000000,
		ColorCode:   1,
		Description: "test client for the go-homebrew project",
		URL:         "https://github.com/BrandMeister/go-homebrew",
	}
	c, err := homebrew.NewClient(config, *addr, *pass)
	if err != nil {
		log.Fatalln(err)
	}

	f := make(chan *homebrew.DMRData)
	go handle(f)
	log.Fatalln(c.ListenAndServe(f))
}
