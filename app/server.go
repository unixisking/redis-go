package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	// RESP Data Types
	RespSimpleString = '+'
	RespBulkString   = '$'
	RespArray        = '*'

	// RESP Commands
	CmdPing = "PING"
	CmdEcho = "ECHO"
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
		connection, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		go handle_connection(connection)
	}
}

func handle_connection(connection net.Conn) {
	defer connection.Close()

	buffer := make([]byte, 1024)

	reader := bufio.NewReader(connection)
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

		command, data := func() (string, string) {
			parsed := parseCommand(string(receivedData))
			if len(parsed) == 1 {
				parsed = append(parsed, "")
			}

			return parsed[0], parsed[1]
		}()

		switch command {
		case CmdPing:
			response := "+PONG\r\n"
			_, writeErr := connection.Write([]byte(response))
			if writeErr != nil {
				fmt.Println("Error write stream to client", err)
				break
			}
		case CmdEcho:
			response := "$" + strconv.Itoa(len(data)) + "\r\n" + data + "\r\n"
			_, writeErr := connection.Write([]byte(response))
			if writeErr != nil {
				fmt.Println("Error write stream to client", err)
				break
			}
		}
	}
}

func parseCommand(requestString string) []string {
	re := regexp.MustCompile(`\d+\r\n([A-Za-z ]+)`)

	numElements, _ := strconv.Atoi(string(requestString[0]))

	for i := 0; i < numElements; i++ {
		continue
	}

	matches := re.FindAllString(requestString, -1)
	parsedCommand := []string{}

	for _, match := range matches {
		literal := strings.Split(match, "\r\n")[1]
		parsedCommand = append(parsedCommand, literal)
	}
	return parsedCommand
}
