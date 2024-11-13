package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

type LogEntry struct {
	Timestamp  string `json:"timestamp"`
	ClientAddr string `json:"client_address"`
	EventType  string `json:"event_type"`
	Message    string `json:"message"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: fatcat <port>")
		return
	}

	port := os.Args[1]

	// Create a unique log file name based on the current time
	logFileName := fmt.Sprintf("fatcat_log_%d.json", rand.Intn(100000))
	logFile, err := os.Create(logFileName)
	if err != nil {
		fmt.Println("Error creating log file:", err)
		return
	}
	defer logFile.Close()

	fmt.Println("Server started, logging to:", logFileName)

	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error setting up server:", err)
		return
	}
	defer listen.Close()

	fmt.Println("Listening on port", port)

	for {
		// Accept new clients
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Handle each client in a separate goroutine
		go handleClient(conn, logFile)
	}
}

func handleClient(conn net.Conn, logFile *os.File) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	fmt.Println("Client connected:", clientAddr)

	// Send a welcome message
	welcomeMessage := fmt.Sprintf("Welcome to the server at %s", time.Now().Format("15:04:05"))
	conn.Write([]byte(welcomeMessage + "\n"))

	// Log the connection
	logEntry := LogEntry{
		Timestamp:  time.Now().Format(time.RFC3339),
		ClientAddr: clientAddr,
		EventType:  "connect",
		Message:    "Client connected",
	}
	logToFile(logFile, logEntry)

	// Handle messages from the client
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()

		// Log the message
		logEntry := LogEntry{
			Timestamp:  time.Now().Format(time.RFC3339),
			ClientAddr: clientAddr,
			EventType:  "message",
			Message:    message,
		}
		logToFile(logFile, logEntry)

		// Handle the message, send it back to the correct client or broadcast
		if strings.HasPrefix(message, "@") {
			// Direct message to a specific user
			targetUser := strings.Split(message[1:], " ")[0]
			messageContent := strings.Join(strings.Split(message[1:], " ")[1:], " ")

			// For now, we just echo the message back
			conn.Write([]byte(fmt.Sprintf("[%s] %s: %s\n", time.Now().Format("15:04:05"), targetUser, messageContent)))
		} else {
			// Broadcast message to all clients (you can modify this for targeted delivery)
			conn.Write([]byte(fmt.Sprintf("[%s] %s: %s\n", time.Now().Format("15:04:05"), clientAddr, message)))
		}
	}

	// Handle any scanning errors
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from client:", err)
	}

	// Log the disconnection
	logEntry = LogEntry{
		Timestamp:  time.Now().Format(time.RFC3339),
		ClientAddr: clientAddr,
		EventType:  "disconnect",
		Message:    "Client disconnected",
	}
	logToFile(logFile, logEntry)
}

// Function to log the messages and events to the log file
func logToFile(file *os.File, logEntry LogEntry) {
	logData, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Println("Error marshaling log entry:", err)
		return
	}
	// Write the log entry to the file
	_, err = file.WriteString(string(logData) + "\n")
	if err != nil {
		fmt.Println("Error writing to log file:", err)
	}
}
