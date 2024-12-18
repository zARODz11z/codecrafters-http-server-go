package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit
var dir string

func main() {
	// Extract --directory
	fmt.Printf("os.args: %v\n", os.Args)
	if len(os.Args) > 2 && os.Args[1] == "--directory" {
		dir = os.Args[2]
	}
	fmt.Printf("--directory %s\n", dir)
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
	if err != nil {
		fmt.Println("Error reading request line: ", err.Error())
		return
	}
	fmt.Printf("Request line: %s\n", requestLine)
	parts := strings.Split(strings.TrimSpace(requestLine), " ")
	if len(parts) == 3 {
		fmt.Printf("Method: %s, Path: %s, Test: %s\n", parts[0], parts[1], parts[2])
		path := parts[1]
		if strings.HasPrefix(path, "/echo/") {
			echoText := path[6:] //since /echo/ is 5 chars long
			echoTextLength := strconv.Itoa(len(echoText))
			conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + echoTextLength + "\r\n\r\n" + echoText))
		} else if strings.HasPrefix(path, "/user-agent") {
			// *** Read headers ***
			headers := make(map[string]string)
			for {
				headerLine, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Error reading header line: ", err.Error())
					return
				}
				headerLine = strings.TrimSpace(headerLine)
				if headerLine == "" {
					break // End of headers
				}
				headerParts := strings.SplitN(headerLine, ": ", 2)
				if len(headerParts) == 2 {
					headers[strings.TrimSpace(headerParts[0])] = strings.TrimSpace(headerParts[1])
				}
			}
			userAgent := headers["User-Agent"]
			fmt.Printf("User-Agent: %s\n", userAgent)
			conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + strconv.Itoa(len(userAgent)) + "\r\n\r\n" + userAgent))

		} else if strings.HasPrefix(path, "/files") {
			file_name := strings.Split(path, "/")[2]
			fmt.Printf("file_name: %s\n", file_name)
			full_file_path := filepath.Join(dir, file_name)
			fmt.Printf("full_file_path: %s\n", full_file_path)
			_, err := os.Stat(full_file_path)
			if err != nil {
				if os.IsNotExist(err) {
					conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
				} else {
					fmt.Println("Error checking file existence: ", err.Error())
					conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
				}
				return
			}
			file_contents, err := os.ReadFile(full_file_path)
			if err != nil {
				fmt.Println("Error reading file: ", err.Error())
				conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
				return
			}
			conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: " + strconv.Itoa(len(file_contents)) + "\r\n\r\n" + string(file_contents)))
		} else if path == "/" {
			conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		} else {
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		}
	}
}
