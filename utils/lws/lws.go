package main

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

var (
	port    string
	dirPath string
)

func main() {
	flag.StringVar(&port, "port", "8080", "Port to run the server on")
	flag.StringVar(&dirPath, "dir", ".", "Directory to serve files from")
	flag.Parse()

	if err := validateDir(dirPath); err != nil {
		log.Fatalf("Invalid directory: %v", err)
	}

	fs := http.FileServer(http.Dir(dirPath))
	http.Handle("/", fs)

	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Serving files from %s on http://localhost%s\n", dirPath, addr)
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	commandInterface()
}

func validateDir(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist")
	}
	if err != nil {
		return fmt.Errorf("error accessing directory: %v", err)
	}
	return nil
}

func commandInterface() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter command: ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)
		handleCommand(cmd)
	}
}

func handleCommand(cmd string) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "/help":
		showHelp()
	case "/exit":
		fmt.Println("Exiting server...")
		os.Exit(0)
	case "/list":
		listFiles()
	case "/delete":
		if len(parts) < 2 {
			fmt.Println("Usage: /delete <filename>")
			return
		}
		deleteFile(parts[1])
	case "/upload":
		fmt.Println("Upload functionality not implemented in this version.")
	case "/download":
		if len(parts) < 2 {
			fmt.Println("Usage: /download <filename>")
			return
		}
		downloadFile(parts[1])
	case "/info":
		if len(parts) < 2 {
			fmt.Println("Usage: /info <filename>")
			return
		}
		showFileInfo(parts[1])
	case "/search":
		if len(parts) < 2 {
			fmt.Println("Usage: /search <pattern>")
			return
		}
		searchFiles(parts[1])
	case "/hash":
		if len(parts) < 2 {
			fmt.Println("Usage: /hash <filename>")
			return
		}
		hashFile(parts[1])
	case "/encode":
		if len(parts) < 2 {
			fmt.Println("Usage: /encode <string>")
			return
		}
		encodeString(strings.Join(parts[1:], " "))
	case "/decode":
		if len(parts) < 2 {
			fmt.Println("Usage: /decode <base64_string>")
			return
		}
		decodeString(parts[1])
	case "/sysinfo":
		showSystemInfo()
	case "/diskusage":
		showDiskUsage()
	case "/processes":
		listProcesses()
	case "/network":
		showNetworkInterfaces()
	case "/ping":
		if len(parts) < 2 {
			fmt.Println("Usage: /ping <host>")
			return
		}
		pingHost(parts[1])
	case "/mkdir":
		if len(parts) < 2 {
			fmt.Println("Usage: /mkdir <dirname>")
			return
		}
		createDirectory(parts[1])
	case "/rmdir":
		if len(parts) < 2 {
			fmt.Println("Usage: /rmdir <dirname>")
			return
		}
		removeDirectory(parts[1])
	case "/rename":
		if len(parts) < 3 {
			fmt.Println("Usage: /rename <oldname> <newname>")
			return
		}
		renameFile(parts[1], parts[2])
	case "/compress":
		if len(parts) < 2 {
			fmt.Println("Usage: /compress <filename>")
			return
		}
		compressFile(parts[1])
	case "/decompress":
		if len(parts) < 2 {
			fmt.Println("Usage: /decompress <filename>")
			return
		}
		decompressFile(parts[1])
	case "/tail":
		if len(parts) < 2 {
			fmt.Println("Usage: /tail <filename> [lines]")
			return
		}
		lines := 10
		if len(parts) > 2 {
			lines, _ = strconv.Atoi(parts[2])
		}
		tailFile(parts[1], lines)
	default:
		fmt.Println("Unknown command. Type /help for a list of commands.")
	}
}

func showHelp() {
	commands := []string{
		"/help         : Show this help message",
		"/exit         : Stop the server and exit",
		"/list         : List all files in the served directory",
		"/delete <file>: Delete a specified file",
		"/download <file>: Download a specified file",
		"/info <file>  : Show information about a file",
		"/search <pattern>: Search for files matching a pattern",
		"/hash <file>  : Calculate hash of a file (MD5 and SHA256)",
		"/encode <string>: Encode a string to Base64",
		"/decode <string>: Decode a Base64 string",
		"/sysinfo      : Show system information",
		"/diskusage    : Show disk usage information",
		"/processes    : List running processes",
		"/network      : Show network interface information",
		"/ping <host>  : Ping a host",
		"/mkdir <dir>  : Create a new directory",
		"/rmdir <dir>  : Remove a directory",
		"/rename <old> <new>: Rename a file or directory",
		"/compress <file>: Compress a file (creates .gz)",
		"/decompress <file>: Decompress a .gz file",
		"/tail <file> [lines]: Show the last lines of a file",
	}
	fmt.Println("Available commands:")
	for _, command := range commands {
		fmt.Println(command)
	}
}

func listFiles() {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		log.Printf("Error reading directory: %v\n", err)
		return
	}
	fmt.Println("Files in directory:")
	for _, file := range files {
		fmt.Println(file.Name())
	}
}

func deleteFile(filename string) {
	filePath := filepath.Join(dirPath, filename)
	err := os.Remove(filePath)
	if err != nil {
		log.Printf("Error deleting file: %v\n", err)
		return
	}
	fmt.Printf("Deleted file: %s\n", filename)
}

func downloadFile(filename string) {
	filePath := filepath.Join(dirPath, filename)
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		fmt.Println("File does not exist.")
		return
	}
	fmt.Printf("File %s is ready for download at http://localhost:%s/%s\n", filename, port, filename)
}

func showFileInfo(filename string) {
	filePath := filepath.Join(dirPath, filename)
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		fmt.Println("File does not exist.")
		return
	} else if err != nil {
		log.Printf("Error getting file info: %v\n", err)
		return
	}
	fmt.Printf("File: %s\nSize: %d bytes\nModified: %s\n", info.Name(), info.Size(), info.ModTime())
}

func searchFiles(pattern string) {
	matches, err := filepath.Glob(filepath.Join(dirPath, pattern))
	if err != nil {
		log.Printf("Error searching files: %v\n", err)
		return
	}
	fmt.Printf("Files matching pattern '%s':\n", pattern)
	for _, match := range matches {
		fmt.Println(filepath.Base(match))
	}
}

func hashFile(filename string) {
	filePath := filepath.Join(dirPath, filename)
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	md5Hash := md5.New()
	sha256Hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(md5Hash, sha256Hash), file); err != nil {
		log.Printf("Error calculating hash: %v\n", err)
		return
	}

	fmt.Printf("File: %s\nMD5: %s\nSHA256: %s\n",
		filename,
		hex.EncodeToString(md5Hash.Sum(nil)),
		hex.EncodeToString(sha256Hash.Sum(nil)))
}

func encodeString(s string) {
	encoded := base64.StdEncoding.EncodeToString([]byte(s))
	fmt.Printf("Encoded string: %s\n", encoded)
}

func decodeString(s string) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.Printf("Error decoding string: %v\n", err)
		return
	}
	fmt.Printf("Decoded string: %s\n", string(decoded))
}

func showSystemInfo() {
	v, _ := mem.VirtualMemory()
	c, _ := cpu.Info()
	h, _ := host.Info()
	fmt.Printf("Total memory: %v GB\n", v.Total/1024/1024/1024)
	fmt.Printf("Free memory: %v GB\n", v.Free/1024/1024/1024)
	fmt.Printf("Number of CPUs: %v\n", len(c))
	fmt.Printf("OS: %v\n", h.OS)
	fmt.Printf("Platform: %v\n", h.Platform)
}

func showDiskUsage() {
	usage, err := disk.Usage("/")
	if err != nil {
		log.Printf("Error getting disk usage: %v\n", err)
		return
	}
	fmt.Printf("Total: %v GB\n", usage.Total/1024/1024/1024)
	fmt.Printf("Free: %v GB\n", usage.Free/1024/1024/1024)
	fmt.Printf("Used: %v GB\n", usage.Used/1024/1024/1024)
	fmt.Printf("Usage: %f%%\n", usage.UsedPercent)
}

func listProcesses() {
	processes, err := process.Processes()
	if err != nil {
		log.Printf("Error getting processes: %v\n", err)
		return
	}
	fmt.Println("Running processes:")
	for _, p := range processes[:10] { // Limiting to first 10 for brevity
		name, _ := p.Name()
		fmt.Printf("PID: %d, Name: %s\n", p.Pid, name)
	}
}

func showNetworkInterfaces() {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Error getting network interfaces: %v\n", err)
		return
	}
	for _, iface := range interfaces {
		fmt.Printf("Name: %v\n", iface.Name)
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			fmt.Printf("  Addr: %v\n", addr)
		}
	}
}

func pingHost(host string) {
	cmd := exec.Command("ping", "-n", "4", host)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error pinging host: %v\n", err)
		return
	}
	fmt.Println(string(output))
}

func createDirectory(dirname string) {
	dirPath := filepath.Join(dirPath, dirname)
	err := os.Mkdir(dirPath, 0755)
	if err != nil {
		log.Printf("Error creating directory: %v\n", err)
		return
	}
	fmt.Printf("Created directory: %s\n", dirname)
}

func removeDirectory(dirname string) {
	dirPath := filepath.Join(dirPath, dirname)
	err := os.RemoveAll(dirPath)
	if err != nil {
		log.Printf("Error removing directory: %v\n", err)
		return
	}
	fmt.Printf("Removed directory: %s\n", dirname)
}

func renameFile(oldname, newname string) {
	oldPath := filepath.Join(dirPath, oldname)
	newPath := filepath.Join(dirPath, newname)
	err := os.Rename(oldPath, newPath)
	if err != nil {
		log.Printf("Error renaming file: %v\n", err)
		return
	}
	fmt.Printf("Renamed %s to %s\n", oldname, newname)
}

func compressFile(filename string) {
	// Implementation omitted for brevity
	fmt.Println("Compress functionality not implemented in this version.")
}

func decompressFile(filename string) {
	// Implementation omitted for brevity
	fmt.Println("Decompress functionality not implemented in this version.")
}

func tailFile(filename string, lines int) {
	filePath := filepath.Join(dirPath, filename)
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var buffer []string
	for scanner.Scan() {
		buffer = append(buffer, scanner.Text())
		if len(buffer) > lines {
			buffer = buffer[1:]
		}
	}

	fmt.Printf("Last %d lines of %s:\n", lines, filename)
	for _, line := range buffer {
		fmt.Println(line)
	}
}
