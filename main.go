package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
}

func main() {
	certPath := flag.String("cert", "cert.pem", "cert.pem path")
	keyPath := flag.String("key", "key.pem", "key.pem path")
	flag.Parse()
	log.Println("kt8900copilot server - bg4qbf")
	cert, err := tls.LoadX509KeyPair(*certPath, *keyPath)
	if err != nil {
		log.Fatalln(err)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	listener, err := tls.Listen("tcp", ":8900", config)
	if err != nil {
		log.Fatalln(err)
	}

	appClients = make(map[string]*Client) // init
	loadClients()

	http.HandleFunc("/ws", wsCallback)

	log.Println("server starting ...")
	if err := http.Serve(listener, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
