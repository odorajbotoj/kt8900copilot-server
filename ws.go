package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	CheckOrigin:     func(_ *http.Request) bool { return true },
}

type safeConn struct {
	Conn *websocket.Conn
	lock sync.Mutex
}

func (conn *safeConn) ReadMessage() (int, []byte, error) {
	return conn.Conn.ReadMessage()
}
func (conn *safeConn) WriteMessage(msgType int, data []byte) error {
	conn.lock.Lock()
	defer conn.lock.Unlock()
	return conn.Conn.WriteMessage(msgType, data)
}
func (conn *safeConn) Close() error {
	return conn.Conn.Close()
}

func verifyClient(connId int64, conn *safeConn, c **Client, doneCh chan struct{}, rst *bool) {
	// step 1. check name (for esp32 is mac address)
	msgType, p, err := conn.ReadMessage()
	if err != nil {
		log.Printf("[%d] conn read error: %v\n", connId, err)
		close(doneCh)
		return
	}
	if msgType != websocket.BinaryMessage {
		if err := conn.WriteMessage(websocket.BinaryMessage, []byte{REFUSE, 0, 0}); err != nil { // refuse connection
			log.Printf("[%d] conn refuse: conn write error: %v\n", connId, err)
		}
		log.Printf("[%d] verifying error: invalid message type.", connId)
		close(doneCh)
		return
	}
	var ok bool
	*c, ok = appClients[string(p)]
	if !ok {
		if err := conn.WriteMessage(websocket.BinaryMessage, []byte{REFUSE, 0, 0}); err != nil { // refuse connection
			log.Printf("[%d] conn refuse: conn write error: %v\n", connId, err)
		}
		log.Printf("[%d] verifying error: invalid client: %s.", connId, string(p))
		close(doneCh)
		return
	}
	log.Printf("[%d] got client id: %s\n", connId, string(p))
	if (*c).activated {
		if err := conn.WriteMessage(websocket.BinaryMessage, []byte{BUSY, 0, 0}); err != nil { // connection busy
			log.Printf("[%d] conn busy: conn write error: %v\n", connId, err)
		}
		log.Printf("[%d] verifying error: client busy.", connId)
		close(doneCh)
		return
	}
	(*c).activated = true
	// step 2. ask for verify
	r1 := rand.New(rand.NewSource(time.Now().UnixMicro())).Uint64()
	r2 := rand.New(rand.NewSource(time.Now().UnixMicro())).Uint64()
	verifyBytes := []byte{VERIFY, 0x10, 0x00,
		byte(r1 & 0xFF), byte(r1 >> 8 & 0xFF), byte(r1 >> 16 & 0xFF), byte(r1 >> 24 & 0xFF),
		byte(r1 >> 32 & 0xFF), byte(r1 >> 40 & 0xFF), byte(r1 >> 48 & 0xFF), byte(r1 >> 56 & 0xFF),
		byte(r2 >> 56 & 0xFF), byte(r2 >> 48 & 0xFF), byte(r2 >> 40 & 0xFF), byte(r2 >> 32 & 0xFF),
		byte(r2 >> 24 & 0xFF), byte(r2 >> 16 & 0xFF), byte(r2 >> 8 & 0xFF), byte(r2 & 0xFF),
	}
	if err := conn.WriteMessage(websocket.BinaryMessage, verifyBytes); err != nil { // verify request
		log.Printf("[%d] verifying error: failed to send verifying codes, error: %v", connId, err)
		close(doneCh)
		return
	}
	var beforeMd5 []byte
	beforeMd5 = append(beforeMd5, (*c).ClientId...)
	beforeMd5 = append(beforeMd5, ':')
	beforeMd5 = append(beforeMd5, verifyBytes[3:]...)
	beforeMd5 = append(beforeMd5, ':')
	beforeMd5 = append(beforeMd5, (*c).Passkey...)
	afterMd5 := md5.Sum(beforeMd5)
	msgType, p, err = conn.ReadMessage()
	if err != nil {
		log.Printf("[%d] conn read error: %v\n", connId, err)
		close(doneCh)
		return
	}
	if msgType != websocket.BinaryMessage || len(p) != 16 {
		if err := conn.WriteMessage(websocket.BinaryMessage, []byte{REFUSE, 0, 0}); err != nil { // refuse connection
			log.Printf("[%d] conn refuse: conn write error: %v\n", connId, err)
		}
		log.Printf("[%d] verifying error: invalid verifying response.", connId)
		close(doneCh)
		return
	}
	if !bytes.Equal(afterMd5[:], p) {
		if err := conn.WriteMessage(websocket.BinaryMessage, []byte{REFUSE, 0, 0}); err != nil { // refuse connection
			log.Printf("[%d] conn refuse: conn write error: %v\n", connId, err)
		}
		log.Printf("[%d] verifying error: unequal verifying response.", connId)
		close(doneCh)
		return
	}
	for len((*c).chanToWs) != 0 {
		<-(*c).chanToWs // clear the channel
	}
	close(doneCh)
	*rst = true
	if err := conn.WriteMessage(websocket.BinaryMessage, []byte{VERIFIED, 0, 0}); err != nil {
		log.Printf("[%d] device verified: conn write error: %v\n", connId, err)
	}
	log.Printf("[%d] device verified: %s\n", connId, (*c).ClientName)
}

func wsCallback(w http.ResponseWriter, r *http.Request) {
	var c *Client
	connId := time.Now().UnixMilli()
	wsc, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[%d] failed to upgrade to websocket: %v\n", connId, err)
		return
	}
	conn := &safeConn{wsc, sync.Mutex{}}
	defer func() {
		log.Printf("[%d] connection closed.\n", connId)
		conn.Close()
		b := []byte{OFFLINE, uint8(len(c.ClientName) & 0xFF), uint8((len(c.ClientName) >> 8) & 0xFF)}
		if c != nil {
			c.chanFromWs <- append(b, c.ClientName...)
			c.activated = false
		}
	}()
	log.Printf("[%d] client connected.\n", connId)

	// verify connection
	doneCh := make(chan struct{})
	rst := false
	go verifyClient(connId, conn, &c, doneCh, &rst)

	select {
	case <-time.After(5 * time.Second):
		if err := conn.WriteMessage(websocket.BinaryMessage, []byte{REFUSE, 0, 0}); err != nil { // refuse connection
			log.Printf("[%d] conn refuse: conn write error: %v\n", connId, err)
		}
		log.Printf("[%d] connection closed: client verifying timeout\n", connId)
		return
	case <-doneCh: // main loop
		if !rst {
			log.Printf("[%d] connection closed: verifying failed or connection busy.\n", connId)
			return
		}
		b := []byte{ONLINE, uint8(len(c.ClientName) & 0xFF), uint8((len(c.ClientName) >> 8) & 0xFF)}
		c.chanFromWs <- append(b, c.ClientName...)
		errChan := make(chan error)
		// set ping/pong
		if c.ClientType == ClientTypeESP32 {
			wsc.SetReadDeadline(time.Now().Add(60 * time.Second))
			wsc.SetPingHandler(func(appData string) error {
				if err := conn.WriteMessage(websocket.PongMessage, nil); err != nil {
					return fmt.Errorf("cannot write pong message to conn: %v.", err)
				}
				return wsc.SetReadDeadline(time.Now().Add(60 * time.Second))
			})
		}
		// read from conn
		go func() {
			for {
				msgType, p, err := conn.ReadMessage()
				if err != nil {
					errChan <- err
					return
				}
				if msgType == websocket.BinaryMessage && len(p) != 0 {
					c.chanFromWs <- p
				}
			}
		}()
		// process data
		for {
			select {
			case b := <-c.chanToWs:
				if err := conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
					log.Printf("[%d] conn write error: %v\n", connId, err)
					continue
				}
			case err := <-errChan:
				log.Printf("[%d] websocket reading error: %v\n", connId, err)
				return
			}
		}
	}
}
