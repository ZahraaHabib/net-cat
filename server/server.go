package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var clients = make(map[net.Conn]string)
var mutex = &sync.Mutex{}
var messages = make([]string, 0)

func main() {
	// Get the port number from the command line and handling errors
	var port int
	if len(os.Args) == 2 {
		portStr := os.Args[1]
		var err error
		port, err = strconv.Atoi(portStr)
		if err != nil {
			fmt.Println("error converting port number")
			return
		}
	} else if len(os.Args) < 2 {
		port = 8989
	} else {
		fmt.Println("too many arguments")
		return
	}

	// Listen for incoming connections
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer listener.Close()

	fmt.Printf("Server is listening on port %d\n", port)

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		// Check if the maximum number of clients has been reached
		if len(clients) >= 10 {
			conn.Write([]byte("Maximum number of clients reached. Please try again later.\n"))
			conn.Close()
			continue
		}

		// Handle client connection in a goroutine
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	fmt.Println("New client is connected to the server")

	// Send the welcome logo to the client
	welcomeLogo, err := os.ReadFile("welcom.txt")
	if err != nil {
		fmt.Println("Error reading welcome logo:", err)
		return
	}
	logoWithPrompt := append(welcomeLogo, []byte("[ENTER YOUR NAME]: ")...)
	_, err = conn.Write(logoWithPrompt)
	if err != nil {
		fmt.Println("Error sending welcome logo:", err)
		return
	}

	// Read the client's name
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "wsarecv: An existing connection was forcibly closed by the remote host." {
			fmt.Println("Client terminated the connection")
		} else {
			fmt.Println("Error:", err)
		}
		return
	}
	clientName := strings.TrimSpace(string(buffer[:n]))

	// Validate the client's name
	for clientName == "" {
		_, err := conn.Write([]byte("Name cannot be empty! [ENTER YOUR NAME]: "))
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		n, err := conn.Read(buffer)
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "wsarecv: An existing connection was forcibly closed by the remote host." {
				fmt.Println("Client terminated the connection")
			} else {
				fmt.Println("Error:", err)
			}
			return
		}
		clientName = strings.TrimSpace(string(buffer[:n]))
	}

	// Add the client to the clients map
	mutex.Lock()
	clients[conn] = clientName
	mutex.Unlock()

	// Send previous messages to the new client
	for _, message := range messages {
		_, err := conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Error sending previous message to", clientName, ":", err)
			return
		}
	}

	// Send the "clientName has joined our chat..." message to all clients
	broadcast(clientName+" has joined our chat...\n", "")

	// Read data from the client
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "wsarecv: An existing connection was forcibly closed by the remote host." {
				fmt.Printf("\n%s has left our chat...\n", clientName)
				broadcast(clientName+" has left our chat...\n", "")
			} else {
				fmt.Println("Error:", err)
			}
			return
		}
		message := strings.TrimSpace(string(buffer[:n]))
		if message == "" {
			continue
		}
		broadcast(message, clientName)
	}
}

func broadcast(message string, clientName string) {
	if message == "" {
		return
	}

	mutex.Lock()
	if strings.Contains(message, "has joined our chat...") || strings.Contains(message, "has left our chat...") {
		messages = append(messages, message)
	} else {
		messages = append(messages, fmt.Sprintf("[%s][%s]: %s\n", time.Now().Format("2006-01-02 15:04:05"), clientName, message))
	}
	for conn, name := range clients {
		var msg string
		if strings.Contains(message, "has joined our chat...") || strings.Contains(message, "has left our chat...") {
			msg = message
		} else {
			msg = fmt.Sprintf("[%s][%s]: %s\n", time.Now().Format("2006-01-02 15:04:05"), clientName, message)
		}
		_, err := conn.Write([]byte(msg))
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "wsarecv: An existing connection was forcibly closed by the remote host." {
				delete(clients, conn)
			} else {
				fmt.Println("connction was terminated by", name)
				delete(clients, conn)
			}
		}
	}
	mutex.Unlock()
}
