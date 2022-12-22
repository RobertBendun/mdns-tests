package main

import (
	"flag"
	"fmt"
	"github.com/hashicorp/mdns"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"
	"io"
)

var (
	serviceName = "_test_message._tcp"
	message string
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

	host, err := os.Hostname()
	if err != nil {
		log.Fatalf("hostname: %v\n", err)
	}

	info := []string { "Example HTTP service" }
	service, err := mdns.NewMDNSService(host, serviceName, "", "", port, nil, info)
	if err != nil {
		log.Fatalf("new mDNS service: %v\n", err)
	}

	server, err := mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		log.Fatalf("new server: %v\n", err)
	}
	defer server.Shutdown()

	select {}
}

func connect(service *mdns.ServiceEntry) {
	target := fmt.Sprintf("http://%s:%d", service.AddrV4, service.Port)
	response, err := http.Get(target)
	if err != nil {
		log.Println("failed to connect to", target, "due to", err)
		return
	}

	fmt.Printf("%s: ", target)
	_, err = io.Copy(os.Stdout, response.Body)
	if err != nil {
		log.Println("failed to read response:", err)
		return
	}
}

func client() {
	services := make(chan *mdns.ServiceEntry, 32)
	go func() {
		mdns.Lookup(serviceName, services)
		close(services)
	}()

	for service := range services {
		connect(service)
	}
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
