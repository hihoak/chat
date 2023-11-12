package chat

import (
	"fmt"
	"log"
	"math/rand"
	"net"
)

const (
	defaultNick = "anonymous"
)

type User struct {
	ID          int
	DoneChan    chan struct{}
	Conn        net.Conn
	Server      *Server
	Nickname    string
	CurrentRoom *Room
	Commands    chan Command
}

func NewUser(conn net.Conn, server *Server) *User {
	return &User{
		ID:          rand.Int(),
		DoneChan:    make(chan struct{}),
		Conn:        conn,
		Server:      server,
		Nickname:    defaultNick,
		CurrentRoom: nil,
		Commands:    server.Commands,
	}
}

func (u *User) StartSession() error {
	_, err := reply(u.Conn, "hello, lets get started!")
	if err != nil {
		return fmt.Errorf("failed to send initial words: %w", err)
	}
	buffer := make([]byte, 2048)
	defer func() {
		if closeErr := u.Conn.Close(); closeErr != nil {
			log.Println("ERROR: failed to close connection for user: ", closeErr)
		}
	}()
mainLoop:
	for {
		select {
		case <-u.DoneChan:
			break mainLoop
		default:
		}

		if _, promptErr := printPrompt(u.Conn); promptErr != nil {
			log.Println("ERROR: failed to print prompt: ", promptErr)
		}
		n, readErr := u.Conn.Read(buffer)
		if readErr != nil {
			return fmt.Errorf("failed to read incoming messages from user: %w", readErr)
		}

		cmd, parseCmdErr := ParseCommand(buffer[:n], u)
		if parseCmdErr != nil {
			_, writeErr := reply(u.Conn, fmt.Sprintf("invalid command: %s", parseCmdErr))
			if writeErr != nil {
				return fmt.Errorf("failed to send parse cmd error: %w", writeErr)
			}
			continue
		}

		if cmdErr := cmd.Run(); cmdErr != nil {
			_, writeErr := reply(u.Conn, fmt.Sprintf("invalid command: %s", parseCmdErr))
			if writeErr != nil {
				log.Println("failed to send cmd run error report to user: ", writeErr)
				continue
			}
		}
	}

	return nil
}
