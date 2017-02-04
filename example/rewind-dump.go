package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"

	"github.com/brandmeister/go-brandmeister/rewind"
)

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

	c.ApplicationCallback = func(dataType uint16, parsed interface{}) {
		switch {
		case dataType >= rewind.TypeDMRDataBase && dataType < rewind.TypeDMRAudioBase:
			log.Printf("received: DMR data type %d\n", dataType-rewind.TypeDMRDataBase)
		case dataType >= rewind.TypeDMRAudioBase && dataType < rewind.TypeDMREmbeddedData:
			log.Println("received: DMR audio frame")
		case dataType == rewind.TypeDMREmbeddedData:
			log.Println("received: DMR embedded data")
		case dataType == rewind.TypeSuperHeader:
			log.Printf("received: super header: %+v\n", parsed)
		}
	}

	for _, s := range strings.Split(*subs, ",") {
		s = strings.TrimSpace(s)
		if tg, err := strconv.ParseUint(s, 10, 32); err != nil {
			log.Fatalf("error parsing %q: %v\n", s, err)
		} else if tg > 0 {
			c.Subscriptions[uint32(tg)] = rewind.GroupVoice
		}
	}

	log.Fatalln(c.ListenAndServe())
}
