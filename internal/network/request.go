package network

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

func Fetch(url string) (string, error) {
	isHTTPS := strings.HasPrefix(url, "https://")
	address := strings.TrimPrefix(url, "https://")
	address = strings.TrimPrefix(address, "http://")

	host, path, found := strings.Cut(address, "/")
	if !found {
		path = "/"
	} else {
		path = "/" + path
	}

	var conn net.Conn
	var err error

	if isHTTPS {
		conn, err = tls.Dial("tcp", host+":443", &tls.Config{InsecureSkipVerify: true})
	} else {
		conn, err = net.Dial("tcp", host+":80")
	}

	if err != nil {
		return "", fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	fmt.Fprintf(conn, "GET %s HTTP/1.1\r\n", path)
	fmt.Fprintf(conn, "Host: %s\r\n", host)
	fmt.Fprintf(conn, "User-Agent: Go-Browser-Project/1.0\r\n")
	fmt.Fprintf(conn, "Connection: close\r\n\r\n")

	reader := bufio.NewReader(conn)
	isChunked := false
	reader.ReadString('\n')

	for {
		line, _ := reader.ReadString('\n')
		if line == "\r\n" || line == "" {
			break
		}
		if strings.Contains(strings.ToLower(line), "transfer-encoding: chunked") {
			isChunked = true
		}
	}

	var result strings.Builder
	if isChunked {
		for {
			line, _ := reader.ReadString('\n')
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			size, err := strconv.ParseInt(line, 16, 64)
			if err != nil || size == 0 {
				break
			}

			chunk := make([]byte, size)
			io.ReadFull(reader, chunk)
			result.Write(chunk)
			reader.Discard(2)
		}
	} else {
		io.Copy(&result, reader)
	}

	return result.String(), nil
}
