package main

import (
	"bytes"
	"crypto/md5"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func verifyClient(connId int64, conn *websocket.Conn, c **Client, doneCh chan<- bool) {
	// step 1. check name (for esp32 is mac address)
	msgType, p, err := conn.ReadMessage()
	if err != nil {
		log.Println(err)
		doneCh <- false
		return
	}
	if msgType != websocket.BinaryMessage {
		if err := conn.WriteMessage(websocket.BinaryMessage, []byte{0x02}); err != nil { // refuse connection
			log.Println(err)
		}
		log.Printf("[%d] verifying error: invalid message type.", connId)
		doneCh <- false
		return
	}
	var ok bool
	*c, ok = appClients[string(p)]
	if !ok {
		if err := conn.WriteMessage(websocket.BinaryMessage, []byte{0x02}); err != nil { // refuse connection
			log.Println(err)
		}
		log.Printf("[%d] verifying error: invalid client.", connId)
		doneCh <- false
		return
	}
	// step 2. ask for verify
	r1 := rand.New(rand.NewSource(time.Now().UnixMicro())).Uint64()
	r2 := rand.New(rand.NewSource(time.Now().UnixMicro())).Uint64()
	verifyBytes := []byte{0x01,
		byte(r1 & 0xFF), byte(r1 >> 8 & 0xFF), byte(r1 >> 16 & 0xFF), byte(r1 >> 24 & 0xFF),
		byte(r1 >> 32 & 0xFF), byte(r1 >> 40 & 0xFF), byte(r1 >> 48 & 0xFF), byte(r1 >> 56 & 0xFF),
		byte(r2 >> 56 & 0xFF), byte(r2 >> 48 & 0xFF), byte(r2 >> 40 & 0xFF), byte(r2 >> 32 & 0xFF),
		byte(r2 >> 24 & 0xFF), byte(r2 >> 16 & 0xFF), byte(r2 >> 8 & 0xFF), byte(r2 & 0xFF),
	}
	if err := conn.WriteMessage(websocket.BinaryMessage, verifyBytes); err != nil { // refuse connection
		log.Println(err)
		log.Printf("[%d] verifying error: failed to send verifying codes.", connId)
		doneCh <- false
		return
	}
	var beforeMd5 []byte
	beforeMd5 = append(beforeMd5, (*c).ClientId...)
	beforeMd5 = append(beforeMd5, ':')
	beforeMd5 = append(beforeMd5, verifyBytes[1:]...)
	beforeMd5 = append(beforeMd5, ':')
	beforeMd5 = append(beforeMd5, (*c).Passkey...)
	afterMd5 := md5.Sum(beforeMd5)
	msgType, p, err = conn.ReadMessage()
	if err != nil {
		log.Println(err)
		doneCh <- false
		return
	}
	if msgType != websocket.BinaryMessage || len(p) != 16 {
		if err := conn.WriteMessage(websocket.BinaryMessage, []byte{0x02}); err != nil { // refuse connection
			log.Println(err)
		}
		log.Printf("[%d] verifying error: invalid verifying response.", connId)
		doneCh <- false
		return
	}
	if !bytes.Equal(afterMd5[:], p) {
		if err := conn.WriteMessage(websocket.BinaryMessage, []byte{0x02}); err != nil { // refuse connection
			log.Println(err)
		}
		log.Printf("[%d] verifying error: unequal verifying response.", connId)
		doneCh <- false
		return
	}
	log.Printf("[%d] device verified: %s\n", connId, (*c).ClientName)
	(*c).activated = true
	doneCh <- true
}

func wsCallback(w http.ResponseWriter, r *http.Request) {
	var c *Client
	connId := time.Now().UnixMilli()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[%d] failed to upgrade to websocket: %v\n", connId, err)
		return
	}
	defer func() {
		if c != nil {
			c.activated = false
		}
		log.Printf("[%d] connection closed.\n", connId)
		conn.Close()
	}()
	log.Printf("[%d] client connected.\n", connId)

	// verify connection
	doneCh := make(chan bool)
	go verifyClient(connId, conn, &c, doneCh)

	select {
	case <-time.After(20 * time.Second):
		if err := conn.WriteMessage(websocket.BinaryMessage, []byte{0x02}); err != nil { // refuse connection
			log.Println(err)
		}
		log.Printf("[%d] connection closed: client verifying timeout\n", connId)
		close(doneCh)
		return
	case ok := <-doneCh: // main loop
		if !ok {
			log.Printf("[%d] connection closed: verifying failed.\n", connId)
			return
		}
		close(doneCh)
		errChan := make(chan error)
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
					log.Println(err)
					continue
				}
			case err := <-errChan:
				log.Printf("[%d] websocket reading error: %v\n", connId, err)
				return
			}
		}
	}
}
