package chat

import (
	"fmt"
	"net"
)

func reply(conn net.Conn, message string) (int, error) {
	return conn.Write([]byte(fmt.Sprintf("%s\n", message)))
}

func printPrompt(conn net.Conn) (int, error) {
	return conn.Write([]byte(" > "))
}
