package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// ServerConfig holds the server configuration details
type ServerConfig struct {
	Port  string            `json:"port"`
	Paths map[string]string `json:"paths"`
}

// Global variables
var (
	config  ServerConfig
	version = "1.0.0"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--serve":
			if len(os.Args) > 2 && os.Args[2] == "." {
				initializeProject()
				startServer()
			} else {
				fmt.Println("Usage: cwf --serve .")
			}
		case "--publish":
			if len(os.Args) > 2 && os.Args[2] == "." {
				publishProject()
			} else {
				fmt.Println("Usage: cwf --publish .")
			}
		case "generate":
			if len(os.Args) > 2 {
				generateFile(os.Args[2])
			} else {
				fmt.Println("Usage: cwf generate <item>")
			}
		case "create-page":
			if len(os.Args) > 2 {
				createPage(os.Args[2])
			} else {
				fmt.Println("Usage: cwf create-page <page-name>")
			}
		case "add-route":
			if len(os.Args) > 3 {
				addRoute(os.Args[2], os.Args[3])
			} else {
				fmt.Println("Usage: cwf add-route <path> <handler>")
			}
		case "list-routes":
			listRoutes()
		case "version":
			printVersion()
		case "help":
			printHelp()
		default:
			fmt.Println("Unknown command. Use 'cwf help' for usage information.")
		}
	} else {
		fmt.Println("Usage: cwf <command> [options]")
		fmt.Println("Use 'cwf help' for more information.")
	}
}

func startServer() {
	// Load configurations
	loadServerConfig()

	// Set up routes
	for path, handler := range config.Paths {
		http.HandleFunc(path, createHandler(handler))
	}

	// Start server
	fmt.Printf("Server starting on port %s...\n", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}

func loadServerConfig() {
	configFile := "config.json"
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error reading config file: %s\n", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatalf("Error parsing config file: %s\n", err)
	}
}

func createHandler(handler string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Handler for %s: %s\n", r.URL.Path, handler)
	}
}

func initializeProject() {
	configFile := "config.json"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Create default config if it doesn't exist
		defaultConfig := ServerConfig{
			Port: "8080",
			Paths: map[string]string{
				"/": "index.html",
			},
		}
		data, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			log.Fatalf("Error creating default config: %s\n", err)
		}
		if err := ioutil.WriteFile(configFile, data, 0644); err != nil {
			log.Fatalf("Error writing config file: %s\n", err)
		}
		fmt.Printf("Initialized project with default config at %s\n", configFile)
	} else {
		fmt.Printf("Config file already exists at %s\n", configFile)
	}
}

func publishProject() {
	fmt.Println("Publishing project...")
	// Placeholder for project publishing logic
}

func generateFile(item string) {
	fmt.Printf("Generating %s file...\n", item)
	var filePath string
	switch item {
	case "html":
		filePath = "index.html"
		err := ioutil.WriteFile(filePath, []byte("<html><body><h1>Hello, World!</h1></body></html>"), 0644)
		if err != nil {
			fmt.Println("Error creating HTML file:", err)
		}
	case "css":
		filePath = "styles.css"
		err := ioutil.WriteFile(filePath, []byte("body { font-family: Arial; }"), 0644)
		if err != nil {
			fmt.Println("Error creating CSS file:", err)
		}
	case "js":
		filePath = "script.js"
		err := ioutil.WriteFile(filePath, []byte("console.log('Hello, World!');"), 0644)
		if err != nil {
			fmt.Println("Error creating JS file:", err)
		}
	default:
		fmt.Println("Unsupported item type. Supported types: html, css, js.")
		return
	}
	fmt.Printf("Created %s file at %s\n", item, filePath)
}

func createPage(pageName string) {
	pagePath := filepath.Join(".", pageName+".html")
	if _, err := os.Create(pagePath); err != nil {
		fmt.Printf("Error creating page: %s\n", err)
		return
	}
	fmt.Printf("Created new page: %s\n", pagePath)
}

func addRoute(path, handler string) {
	config.Paths[path] = handler
	fmt.Printf("Added route: %s -> %s\n", path, handler)
}

func listRoutes() {
	fmt.Println("Listing all routes:")
	for path, handler := range config.Paths {
		fmt.Printf("Path: %s, Handler: %s\n", path, handler)
	}
}

func printVersion() {
	fmt.Printf("CWF Version: %s\n", version)
}

func printHelp() {
	fmt.Println("Usage: cwf <command> [options]")
	fmt.Println("Commands:")
	fmt.Println("--serve .             Start the server with current directory as the project.")
	fmt.Println("--publish .           Publish the project.")
	fmt.Println("generate <item>      Generate a new file (html, css, js).")
	fmt.Println("create-page <name>   Create a new HTML page.")
	fmt.Println("add-route <path> <handler>  Add a new route.")
	fmt.Println("list-routes          List all registered routes.")
	fmt.Println("version              Show the version of CWF.")
	fmt.Println("help                 Show help information.")
}
