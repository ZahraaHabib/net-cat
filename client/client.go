package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("[USAGE]: go run . $port $hostname")
		return
	}

	portStr := os.Args[1]
	hostname := os.Args[2]

	// Validate port number
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		fmt.Println("[USAGE]: go run . $port $hostname")
		return
	}

	// Connect to the server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer conn.Close()

	// Read the server's message
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Print(string(buffer[:n]))

	// Send the client's name
	reader := bufio.NewReader(os.Stdin)
	clientName, _ := reader.ReadString('\n')
	clientName = strings.TrimSpace(clientName)
	_, err = conn.Write([]byte(clientName))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Start a goroutine to read messages from the server
	go func() {
		for {
			buffer := make([]byte, 1024)
			n, err := conn.Read(buffer)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			fmt.Print(string(buffer[:n]))
		}
	}()

	// Read input from the user and send it to the server
	reader = bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		_, err := conn.Write([]byte(input))
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}
