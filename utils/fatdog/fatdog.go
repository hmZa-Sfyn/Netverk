package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: fatdog <server-ip> <port>")
		return
	}

	serverAddr := os.Args[1]
	port := os.Args[2]

	// Establish connection to the server
	conn, err := net.Dial("tcp", serverAddr+":"+port)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to", serverAddr+":"+port)

	// Handle user input and send it to the server
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("Enter message: ")
			scanner.Scan()
			input := scanner.Text()

			if input == "exit" {
				return
			}

			// Send the message to the server
			_, err := conn.Write([]byte(input + "\n"))
			if err != nil {
				fmt.Println("Error sending data:", err)
				return
			}
		}
	}()

	// Receive data from the server
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading data:", err)
			return
		}
		fmt.Printf("Received: %s", string(buffer[:n]))
	}
}
