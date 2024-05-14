package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// HandlerFunc represents a function that can handle incoming connections.
type HandlerFunc func(net.Conn)

// Route represents a TCP route and its associated handler function.
type Route struct {
	Pattern string
	Handler HandlerFunc
}

// TCPHandler is responsible for routing incoming connections based on their patterns.
type TCPHandler struct {
	routes []Route
}

// NewTCPHandler creates a new TCPHandler.
func NewTCPHandler() *TCPHandler {
	return &TCPHandler{}
}

// HandleFunc registers a handler for a specific pattern.
func (h *TCPHandler) HandleFunc(pattern string, handler HandlerFunc) {
	h.routes = append(h.routes, Route{Pattern: pattern, Handler: handler})
}

// ServeTCP accepts incoming connections and routes them to the appropriate handler.
func (h *TCPHandler) ServeTCP(listener net.Listener) error {
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go h.handleConnection(conn)
	}
}

// handleConnection routes an incoming connection to the appropriate handler based on its pattern.
func (h *TCPHandler) handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	// Create a reader to read from the connection
	reader := bufio.NewReader(conn)

	// Read the request line
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading request line:", err.Error())
		return
	}

	// Parse the request line
	parts := strings.Fields(requestLine)
	if len(parts) != 3 {
		fmt.Println("Malformed request line:", requestLine)
		return
	}
	method, path, _ := parts[0], parts[1], parts[2]

	// Read the headers
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil || line == "\r\n" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	// Read the body if content length is specified
	var body string
	contentLengthStr, ok := headers["Content-Length"]
	if ok {
		contentLength := 0
		fmt.Sscanf(contentLengthStr, "%d", &contentLength)
		bodyBytes := make([]byte, contentLength)
		_, err := reader.Read(bodyBytes)
		if err != nil {
			fmt.Println("Error reading body:", err.Error())
			return
		}
		body = string(bodyBytes)
	}

	// Print the received request
	fmt.Println("Received request:")
	fmt.Println("Method:", method)
	fmt.Println("Path:", path)
	fmt.Println("Headers:", headers)
	fmt.Println("Body:", body)

	// Find a matching route and handle the request
	for _, route := range h.routes {
		if route.Pattern == path {
			route.Handler(conn)
			return
		}
	}

	conn.Write([]byte("HTTP/1.1 400 OK\r\nContent-Type: text/html\r\n\r\n<h1>Route Not Found</h1>"))

}

func main() {
	// Create a new TCPHandler instance
	handler := NewTCPHandler()

	// Register handler functions for specific routes
	handler.HandleFunc("/", func(conn net.Conn) {
		// Send an HTTP response
		response := "HTTP/1.1 200 OK\r\nContent-Type: text/html\r\n\r\n<h1>Hello Dev</h1>"
		conn.Write([]byte(response))
	})

	// Start listening on port 8080
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer listener.Close()
	fmt.Println("Server is listening on port 8080")

	// Serve incoming connections using the TCPHandler
	err = handler.ServeTCP(listener)
	if err != nil {
		fmt.Println("Error serving:", err.Error())
		return
	}
}
