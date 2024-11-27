package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	requestLine, err := reader.ReadString('\n')
	// fmt.Println(requestLine)
	if err != nil {
		fmt.Println("Error reading request line: ", err.Error())
		return
	}

	parts := strings.Split(strings.TrimSpace(requestLine), " ")
	if len(parts) == 3 {
		fmt.Printf("Method: %s, Path: %s\n", parts[0], parts[1])
		path := parts[1]
		if strings.HasPrefix(path, "/echo/") {
			echoText := path[6:]
			echoTextLength := strconv.Itoa(len(echoText))
			conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + echoTextLength + "\r\n\r\n" + echoText))
		} else if path == "/" {
			conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		} else {
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		}
	}
}
