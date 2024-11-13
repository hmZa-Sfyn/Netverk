package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

var (
	host string
	port string
)

// Connect to TCP server and start chatting
func startClient() {
	conn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		log.Fatalf("Unable to connect to server: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connected to the server!")

	// Go routine to listen for server messages
	go func() {
		for {
			message, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				log.Println("Server connection lost.")
				return
			}
			fmt.Print("\n" + message)
		}
	}()

	// Main routine to send messages to server
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if scanner.Scan() {
			text := scanner.Text()
			fmt.Fprintln(conn, text)
		}
	}
}

func main() {
	// Command-line options
	flag.StringVar(&host, "host", "127.0.0.1", "Host to connect to")
	flag.StringVar(&port, "port", "8080", "Port to connect to")
	flag.Parse()

	fmt.Printf("Connecting to server at %s:%s...\n", host, port)
	startClient()
}
