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
	"time"
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

type item struct {
	value      string
	lastAccess int64
	ttl        int
}

type DB struct {
	items map[string]item
}

type CommandData struct {
	Command string
	Args    []string
}

func main() {
	fmt.Println("Server listening on port 6379...")

	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer listener.Close()

	db := DB{items: make(map[string]item)}

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		go handle_connection(connection, &db)
	}
}

func handle_connection(connection net.Conn, db *DB) {
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

		switch {
		case strings.EqualFold(CmdPing, data.Command):
			response := "+PONG\r\n"
			_, writeErr := connection.Write([]byte(response))
			if writeErr != nil {
				fmt.Println("Error write stream to client", err)
				break
			}

		case strings.EqualFold(CmdEcho, data.Command):
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

		case strings.EqualFold(CmdSet, data.Command):
			if len(data.Args) < 2 {
				fmt.Println("Not enough arguments provided")
				break
			}
			key := data.Args[0]
			val := data.Args[1]
			response := "+OK\r\n"
			// db.items[key] = item{value: val}
			kvSet(key, val, db)
			_, writeErr := connection.Write([]byte(response))
			if writeErr != nil {
				fmt.Println("Error write stream to client", err)
				break
			}
		case strings.EqualFold(CmdGet, data.Command):
			it := kvGet(data.Args[0], db)
			if it != nil {
				response := "$" + strconv.Itoa(len(it.value)) + "\r\n" + it.value + "\r\n"
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
	re := regexp.MustCompile(`\d+\r\n([A-Za-z0-9 ]+)`)

	matches := re.FindAllString(requestString, -1)
	fmt.Printf("matches: %q", matches)
	parsedCommand := []string{}

	for _, match := range matches {
		literal := strings.Split(match, "\r\n")[1]
		parsedCommand = append(parsedCommand, literal)
	}
	return parsedCommand
}

/**
* Sets a value given a key, if it already exists update lastAccess timestamp
* */
func kvSet(key string, val string, db *DB) {
	it, ok := db.items[key]
	if !ok {
		it = item{value: val}
		db.items[key] = it
	}
	it.lastAccess = time.Now().Unix()
}

/**
* gets a value given a key, return nil if it doesn't exist
* */
func kvGet(key string, db *DB) *item {
	it, ok := db.items[key]
	if !ok {
		return nil
	}
	it.lastAccess = time.Now().Unix()
	return &it
}
