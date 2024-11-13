package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	port           string
	mu             sync.Mutex
	clients        = make(map[net.Conn]string) // Connected clients
	clientLog      = make(map[string][]string) // Logs for each client
	totalMessages  int                         // Count of total messages sent
	startTime      = time.Now()                // Server start time
	messages       = make(chan string)         // Channel for broadcast messages
	blockedUsers   = make(map[string]bool)     // Blocked users
	blockedIPs     = make(map[string]bool)     // Blocked IPs
	whitelistedIPs = make(map[string]bool)     // Whitelisted IPs
)

type clientInfo struct {
	username    string
	connectTime time.Time
}

// Active client connections with metadata
var clientData = make(map[net.Conn]clientInfo)

// Broadcast a message to all clients
func broadcastMessage(message string) {
	mu.Lock()
	defer mu.Unlock()
	for client := range clients {
		fmt.Fprintln(client, message)
	}
}

// Handle client connections and commands
func handleClient(conn net.Conn) {
	defer conn.Close()

	// Check if IP is blocked or not whitelisted
	ip := conn.RemoteAddr().(*net.TCPAddr).IP.String()
	if blockedIPs[ip] || (len(whitelistedIPs) > 0 && !whitelistedIPs[ip]) {
		conn.Write([]byte("Your IP is blocked or not whitelisted.\n"))
		return
	}

	reader := bufio.NewReader(conn)
	conn.Write([]byte("Enter your name: "))
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	// Check if user is blocked
	if blockedUsers[name] {
		conn.Write([]byte("(!) You are blocked from this server.\n"))
		fmt.Printf(fmt.Sprintf("* `%s` is blocked but still tryed to login at %d", name, time.Now().Format(time.RFC1123)))
		return
	}

	mu.Lock()
	clients[conn] = name
	clientData[conn] = clientInfo{username: name, connectTime: time.Now()}
	clientLog[name] = append(clientLog[name], fmt.Sprintf("%s joined at %s", name, time.Now().Format(time.RFC1123)))
	mu.Unlock()

	// Welcome message
	welcomeMsg := fmt.Sprintf("\n* `%s` has joined the chat!", name)
	messages <- welcomeMsg

	fmt.Printf(fmt.Sprintf("* `%s` joined at %s", name, time.Now().Format(time.RFC1123)))

	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("\n* `%s` disconnected.\n", name)
			mu.Lock()
			delete(clients, conn)
			delete(clientData, conn)
			mu.Unlock()
			messages <- fmt.Sprintf("* `%s` has left the chat.", name)
			return
		}
		msg = strings.TrimSpace(msg)

		// Handle special commands
		if strings.HasPrefix(msg, "/") {
			handleCommand(msg, conn)
			continue
		}

		fullMsg := fmt.Sprintf("%s: %s", name, msg)
		mu.Lock()
		clientLog[name] = append(clientLog[name], fullMsg)
		mu.Unlock()
		totalMessages++
		messages <- fullMsg
	}
}

// Handle server commands
func handleCommand(command string, conn net.Conn) {
	args := strings.Fields(command)
	switch args[0] {
	// For Hosting Commands
	case "/hostall":
		hostOnAllInterfaces(conn)
	case "/localhost":
		hostOnLocalhost(conn)
	case "/localnet":
		hostOnLocalNetwork(conn)
	// Forensic/Analytics Commands
	case "/stats":
		showStats(conn)
	case "/uptime":
		showUptime(conn)
	case "/msgcount":
		showMessageCount(conn)
	case "/users":
		listUsers(conn)
	case "/kick":
		if len(args) > 1 {
			kickUser(args[1], conn)
		} else {
			fmt.Fprintln(conn, "Usage: /kick $username")
		}
	case "/ban":
		if len(args) > 1 {
			banUser(args[1], conn)
		} else {
			fmt.Fprintln(conn, "Usage: /ban $username")
		}
	case "/unban":
		if len(args) > 1 {
			unbanUser(args[1], conn)
		} else {
			fmt.Fprintln(conn, "Usage: /unban $username")
		}
	case "/log":
		if len(args) > 1 {
			showUserLog(args[1], conn)
		} else {
			fmt.Fprintln(conn, "Usage: /log $username")
		}
	case "/connections":
		showConnections(conn)
	case "/blockip":
		if len(args) > 1 {
			blockIP(args[1], conn)
		} else {
			fmt.Fprintln(conn, "Usage: /blockip $IP")
		}
	case "/whitelistip":
		if len(args) > 1 {
			whitelistIP(args[1], conn)
		} else {
			fmt.Fprintln(conn, "Usage: /whitelistip $IP")
		}
	case "/save":
		saveLogsToFile(conn)
	default:
		fmt.Fprintln(conn, "Unknown command.")
	}
}

// Show server stats
func showStats(conn net.Conn) {
	uptime := time.Since(startTime)
	stats := fmt.Sprintf("Server has been running for %s\nTotal Messages Sent: %d\nTotal Clients Connected: %d\n",
		uptime, totalMessages, len(clients))
	fmt.Fprintln(conn, stats)
}

// Show server uptime
func showUptime(conn net.Conn) {
	uptime := time.Since(startTime)
	fmt.Fprintf(conn, "Server uptime: %s\n", uptime)
}

// Show total message count
func showMessageCount(conn net.Conn) {
	fmt.Fprintf(conn, "Total messages sent: %d\n", totalMessages)
}

// List active users
func listUsers(conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Fprintln(conn, "Active Users:")
	for _, name := range clients {
		fmt.Fprintln(conn, name)
	}
}

// Kick a specific user
func kickUser(username string, conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()
	for client, name := range clients {
		if name == username {
			client.Close()
			delete(clients, client)
			delete(clientData, client)
			messages <- fmt.Sprintf("%s was kicked from the server.", username)
			return
		}
	}
	fmt.Fprintln(conn, "User not found.")
}

// Ban a specific user
func banUser(username string, conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()
	blockedUsers[username] = true
	fmt.Fprintf(conn, "%s has been banned.\n", username)
	// Kick the user if they're currently connected
	for client, name := range clients {
		if name == username {
			client.Close()
			delete(clients, client)
			delete(clientData, client)
			messages <- fmt.Sprintf("%s was banned and kicked from the server.", username)
			return
		}
	}
}

// Unban a specific user
func unbanUser(username string, conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()
	delete(blockedUsers, username)
	fmt.Fprintf(conn, "%s has been unbanned.\n", username)
}

// Show message log of a specific user
func showUserLog(username string, conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()
	if log, exists := clientLog[username]; exists {
		fmt.Fprintf(conn, "Log for %s:\n", username)
		for _, msg := range log {
			fmt.Fprintln(conn, msg)
		}
	} else {
		fmt.Fprintln(conn, "No log found for user.")
	}
}

// Show all active connections with IPs and connection time
func showConnections(conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Fprintln(conn, "Active Connections:")
	for client, info := range clientData {
		addr := client.RemoteAddr().String()
		duration := time.Since(info.connectTime)
		fmt.Fprintf(conn, "%s (%s) - Connected for: %s\n", info.username, addr, duration)
	}
}

// Block an IP from connecting to the server
func blockIP(ip string, conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()
	blockedIPs[ip] = true
	fmt.Fprintf(conn, "IP %s has been blocked.\n", ip)
	// Kick any clients currently connected from this IP
	for client, info := range clientData {
		if client.RemoteAddr().(*net.TCPAddr).IP.String() == ip {
			client.Close()
			delete(clients, client)
			delete(clientData, client)
			messages <- fmt.Sprintf("%s was kicked due to IP block.", info.username)
		}
	}
}

// Whitelist an IP to allow access to the server
func whitelistIP(ip string, conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()
	whitelistedIPs[ip] = true
	delete(blockedIPs, ip) // Remove from blocked list if present
	fmt.Fprintf(conn, "IP %s has been whitelisted.\n", ip)
}

// Save logs to a file
func saveLogsToFile(conn net.Conn) {
	file, err := os.Create("server_logs.txt")
	if err != nil {
		fmt.Fprintln(conn, "Error saving logs.")
		return
	}
	defer file.Close()

	mu.Lock()
	defer mu.Unlock()

	for name, logs := range clientLog {
		file.WriteString(fmt.Sprintf("Log for %s:\n", name))
		for _, msg := range logs {
			file.WriteString(msg + "\n")
		}
		file.WriteString("\n")
	}

	fmt.Fprintln(conn, "Logs saved to server_logs.txt.")
}

// Host the server on all interfaces (0.0.0.0)
func hostOnAllInterfaces(conn net.Conn) {
	port = "0.0.0.0:" + port
	fmt.Fprintf(conn, "Server hosting on 0.0.0.0:%s\n", port)
}

// Host the server on localhost (127.0.0.1)
func hostOnLocalhost(conn net.Conn) {
	port = "127.0.0.1:" + port
	fmt.Fprintf(conn, "Server hosting on 127.0.0.1:%s\n", port)
}

// Host the server on local network (192.168.0.1/24)
func hostOnLocalNetwork(conn net.Conn) {
	port = "192.168.0.1:" + port
	fmt.Fprintf(conn, "Server hosting on 192.168.0.1:%s\n", port)
}

// Start the TCP server
func startServer() {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}
	defer listener.Close()

	go func() {
		for msg := range messages {
			broadcastMessage(msg)
		}
	}()

	log.Printf("Server started on port %s\n", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v\n", err)
			continue
		}
		go handleClient(conn)
	}
}

func main() {
	flag.StringVar(&port, "port", "8080", "Port to listen on")
	flag.Parse()

	startServer()
}
