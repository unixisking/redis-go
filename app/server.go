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
	CmdSet  = "SET"
	CmdGet  = "GET"
)

type DB struct {
	entries map[string]string
}

type CommandData struct {
	Command string
	Args    []string
}

func main() {
	fmt.Println("Server listening on port 6379...")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	db := DB{entries: make(map[string]string)}

	for {
		connection, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		go handle_connection(connection, db)
	}
}

func handle_connection(connection net.Conn, db DB) {
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

		data := func() CommandData {
			parsed := parseCommand(string(receivedData))
			if len(parsed) == 1 {
				parsed = append(parsed, "")
			}

			return CommandData{
				Command: parsed[0],
				Args:    parsed[1:],
			}
		}()

		switch data.Command {
		case CmdPing:
			response := "+PONG\r\n"
			_, writeErr := connection.Write([]byte(response))
			if writeErr != nil {
				fmt.Println("Error write stream to client", err)
				break
			}

		case CmdEcho:
			if len(data.Args) == 0 {
				fmt.Println("No arguments provided")
				break
			}
			argsLength := strconv.Itoa(len(data.Args[0]))
			response := "$" + argsLength + "\r\n" + data.Args[0] + "\r\n"
			_, writeErr := connection.Write([]byte(response))
			if writeErr != nil {
				fmt.Println("Error write stream to client", err)
				break
			}

		case CmdSet:
			if len(data.Args) < 2 {
				fmt.Println("Not enough arguments provided")
				break
			}
			key := data.Args[0]
			val := data.Args[1]
			response := "+OK\r\n"
			db.entries[key] = val
			_, writeErr := connection.Write([]byte(response))
			if writeErr != nil {
				fmt.Println("Error write stream to client", err)
				break
			}
		case CmdGet:
			val, ok := db.entries[data.Args[0]]
			if ok {
				response := "$" + strconv.Itoa(len(val)) + "\r\n" + val + "\r\n"
				_, writeErr := connection.Write([]byte(response))
				if writeErr != nil {
					fmt.Println("Error write stream to client", err)
					break
				}
			} else {
				_, writeErr := connection.Write([]byte("$-1\r\n"))
				if writeErr != nil {
					fmt.Println("Error write stream to client", err)
					break
				}
			}
		default:
			fmt.Println("Unknown command, please check the help manual for suppoorted commands")
			break
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
