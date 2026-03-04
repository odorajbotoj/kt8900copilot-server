package main

import (
	"crypto/tls"
	_ "embed"
	"flag"
	"log"
	"net/http"
)

//go:embed web/talk.html
var simpleClient []byte

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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(simpleClient)
	})

	log.Println("server starting ...")
	if err := http.Serve(listener, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
