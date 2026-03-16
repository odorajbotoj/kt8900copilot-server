package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"slices"
)

const (
	ClientTypeESP32 = iota + 1
	ClientTypeUser
)

type Client struct {
	ClientId          string // id
	ClientType        int    // type
	ClientName        string // name
	ClientMac         string // MAC address (only ESP32)
	OutClientsNames   []string
	Passkey           string // client key
	IgnoreFromChannel []int
	IgnoreFromWs      []int
	activated         bool
	chanFromWs        chan []byte
	chanToWs          chan []byte
	chanIn            chan dataPack
	outClientsPtrs    []*Client
	nowReceiving      string
}

var appClients map[string]*Client

func loadClients() {
	f, err := os.OpenFile("clients.json", os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		log.Fatalf("cannot open clients.json, error: %v\n", err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf("cannot read clients.json, error: %v\n", err)
	}
	err = json.Unmarshal(b, &appClients)
	if err != nil {
		log.Fatalf("cannot load clients.json, error: %v\n", err)
	}
	for _, v := range appClients {
		for _, name := range v.OutClientsNames {
			if c, ok := appClients[name]; ok {
				v.outClientsPtrs = append(v.outClientsPtrs, c)
			}
		}
		go v.initAndServe()
	}
}

func setFrom(c *Client, p *dataPack) {
	if c.ClientType == ClientTypeUser {
		fb := []byte{FROM, uint8(len(p.from) & 0xFF), uint8((len(p.from) >> 8) & 0xFF)}
		if len(c.chanToWs) < cap(c.chanToWs) {
			c.chanToWs <- append(fb, p.from...)
		}
	}
	c.nowReceiving = p.from
}

func (c *Client) initAndServe() {
	c.chanFromWs = make(chan []byte, 4)
	c.chanToWs = make(chan []byte, 4)
	c.chanIn = make(chan dataPack, 4)
	if c.ClientType == ClientTypeESP32 {
		c.ClientId = c.ClientMac
	} else {
		c.ClientId = c.ClientName
	}
	for {
		select {
		case p := <-c.chanIn:
			skip := slices.Contains(c.IgnoreFromChannel, int(p.data[0]))
			if skip {
				continue
			}
			// log.Printf("%s received data from channel %d\n", c.ClientId, p.data[0])
			switch p.data[0] {
			case RX:
				setFrom(c, &p)
				p.data[0] = TX
			case RX_STOP:
				c.nowReceiving = ""
				p.data[0] = TX_STOP
			case TX:
				setFrom(c, &p)
				p.data[0] = RX
			case TX_STOP:
				c.nowReceiving = ""
				p.data[0] = RX_STOP
			}
			if c.nowReceiving == "" || c.nowReceiving == p.from {
				if len(c.chanToWs) < cap(c.chanToWs) {
					c.chanToWs <- p.data
				}
			}
		case b := <-c.chanFromWs:
			skip := slices.Contains(c.IgnoreFromWs, int(b[0]))
			if skip {
				continue
			}
			// log.Printf("%s received data from ws %d\n", c.ClientId, b[0])
			for _, oc := range c.outClientsPtrs {
				if len(oc.chanIn) < cap(oc.chanIn) {
					bCopy := make([]byte, len(b))
					copy(bCopy, b)
					oc.chanIn <- dataPack{from: c.ClientName, data: bCopy}
				}
			}
		}
	}
}
