package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	fmt.Println("Server listening on port 6379...")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		go handle_connection(conn)
	}
}

func handle_connection(c net.Conn) {
	defer c.Close()

	buffer := make([]byte, 1024)

	reader := bufio.NewReader(c)
	for {
		n, err := reader.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading stream", err, n)
			}
			break
		}

		receivedData := buffer[:n]
		fmt.Printf("Received: %q\n", receivedData)

		_, writeErr := c.Write([]byte("+PONG\r\n"))

		if writeErr != nil {
			fmt.Println("Error write stream to client", err)
			break
		}
	}
}
