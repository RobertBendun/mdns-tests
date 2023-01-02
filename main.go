package main

import (
	"flag"
	"fmt"
	"github.com/grandcat/zeroconf"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"context"
)

var (
	serviceName = "_test_message._tcp"
	message     string
)

var port int

func tcpServerWithRandomPort() net.Listener {
	for {
		port = 8000 + rand.Intn(100)

		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			return listener
		}
	}
}

func echoServer() {
	listener := tcpServerWithRandomPort()

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, message)
		}),
	}
	go server.Serve(listener)
}

func server() {
	echoServer()

	server, err := zeroconf.Register("GoZeroconf", serviceName, "local.", port, []string{"txtv=0", "lo=1", "la=2"}, nil)

	if err != nil {
		log.Fatalf("new mDNS service: %v\n", err)
	}
	defer server.Shutdown()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sig:
	}
}

func client() {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err)
		return
	}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			log.Println(entry)
		}
		log.Println("Entries exhausted")
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 10)
	defer cancel()

	err = resolver.Browse(ctx, serviceName, "local.", entries)
	if err != nil {
		log.Fatalln("Failed to browse:", err)
	}
	<-ctx.Done()
}

func main() {
	rand.Seed(time.Now().UnixNano())

	listen := flag.Bool("listen", false, "Create mDNS service that listens for HTTP messages")
	flag.StringVar(&message, "message", "", "Create mDNS service that listens for HTTP messages")
	flag.Parse()

	if *listen {
		if len(message) == 0 {
			fmt.Println("provide message with -message")
			return
		}
		server()
	} else {
		client()
	}
}
