package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

const (
	ClientTypeESP32 = iota + 1
	ClientTypeUser
)

type Client struct {
	ClientId        string // id
	ClientType      int    // type
	ClientName      string // name
	ClientMac       string // MAC address (only ESP32)
	OutClientsNames []string
	Passkey         string // client key
	activated       bool   // is client activated
	chanFromWs      chan []byte
	chanToWs        chan []byte
	chanIn          chan dataPack
	outClientsPtrs  []*Client
	nowReceiving    string
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
		fb := []byte{FROM}
		select {
		case c.chanToWs <- append(fb, p.from...):
		default:
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
			case IMG_UPLOAD:
				setFrom(c, &p)
				p.data[0] = IMG_DOWNLOAD
			case IMG_UPLOAD_STOP:
				c.nowReceiving = ""
				p.data[0] = IMG_DOWNLOAD_STOP
			case IMG_DOWNLOAD:
				setFrom(c, &p)
				p.data[0] = IMG_UPLOAD
			case IMG_DOWNLOAD_STOP:
				c.nowReceiving = ""
				p.data[0] = IMG_UPLOAD_STOP
			}
			if c.nowReceiving == "" || c.nowReceiving == p.from {
				select {
				case c.chanToWs <- p.data:
				default:
				}
			}
		case b := <-c.chanFromWs:
			for _, oc := range c.outClientsPtrs {
				if oc.activated {
					select {
					case oc.chanIn <- dataPack{from: c.ClientId, data: b}:
					default:
					}
				}
			}
		}
	}
}
