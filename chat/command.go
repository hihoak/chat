package chat

import (
	"fmt"
	"log"
	"strings"
)

type CommandName string

const (
	CMD_NICK  = "nick"
	CMD_ROOMS = "rooms"
	CMD_JOIN  = "join"
	CMD_SEND  = "send"
	CMD_QUIT  = "quit"
)

type Command struct {
	Name CommandName
	Args []string
	User *User
}

func ParseCommand(rawData []byte, user *User) (Command, error) {
	strData := string(rawData)
	strData = strings.Trim(strData, "\n\r")
	splitData := strings.Split(strData, " ")
	name := splitData[0]
	var args []string
	if len(splitData) > 1 {
		args = splitData[1:]
	}
	switch strings.TrimSpace(name) {
	case CMD_NICK:
		if len(args) > 1 {
			return Command{}, fmt.Errorf("invalid %q cmd: too much arguments: expect 1 or 0", CMD_NICK)
		}
		return Command{
			Name: CMD_NICK,
			Args: args,
			User: user,
		}, nil
	case CMD_JOIN:
		if len(args) != 1 {
			return Command{}, fmt.Errorf("invalid %q cmd: expect only 1 argument", CMD_JOIN)
		}
		return Command{
			Name: CMD_JOIN,
			Args: args,
			User: user,
		}, nil
	case CMD_ROOMS:
		if len(args) > 0 {
			return Command{}, fmt.Errorf("invalid %q cmd: too much arguments: expect 0", CMD_ROOMS)
		}
		return Command{
			Name: CMD_ROOMS,
			User: user,
		}, nil
	case CMD_SEND:
		if len(args) != 1 {
			return Command{}, fmt.Errorf("invalid %q cmd: expect 1 argument", CMD_SEND)
		}
		return Command{
			Name: CMD_SEND,
			User: user,
			Args: args,
		}, nil
	case CMD_QUIT:
		return Command{
			Name: CMD_QUIT,
			User: user,
		}, nil
	}
	return Command{}, fmt.Errorf("unknown command %q", name)
}

func (c *Command) Run() error {
	switch c.Name {
	case CMD_NICK:
		if len(c.Args) == 0 {
			_, err := reply(c.User.Conn, fmt.Sprintf("your nickname on the server is %q", c.User.Nickname))
			if err != nil {
				return fmt.Errorf("failed to send current user nickname: %w", err)
			}
			return nil
		}
		if len(c.Args) == 1 {
			oldNickname := c.User.Nickname
			c.User.Nickname = c.Args[0]
			_, err := reply(c.User.Conn, fmt.Sprintf("your nickname on the server changed to %q", c.User.Nickname))
			if err != nil {
				c.User.Nickname = oldNickname
				return fmt.Errorf("failed to send reply that nickname changed to user: %w", err)
			}
			return nil
		}
		return fmt.Errorf("unexpected error in %q cmd too much arguments", CMD_NICK)
	case CMD_JOIN:
		if len(c.Args) != 1 {
			return fmt.Errorf("unexpected error: %q cmd accept only 1 argument", CMD_JOIN)
		}
		roomName := c.Args[0]
		room, ok := c.User.Server.Rooms.Get(roomName)
		if !ok {
			var roomCreateErr error
			room, roomCreateErr = c.User.Server.Rooms.Create(roomName)
			if roomCreateErr != nil {
				return fmt.Errorf("failed to create room: %w", roomCreateErr)
			}
		}
		if c.User.CurrentRoom != nil {
			c.User.CurrentRoom.Users.Delete(c.User.ID)
		}
		c.User.CurrentRoom = room
		room.Users.Set(c.User.ID, c.User)
		if _, err := reply(c.User.Conn, room.FormatMessages()); err != nil {
			return fmt.Errorf("failed to list all messages in the room: %w", err)
		}
		return nil
	case CMD_ROOMS:
		_, err := reply(c.User.Conn, fmt.Sprintf("now on the server available following rooms: %v", c.User.Server.Rooms.ListNames()))
		if err != nil {
			return fmt.Errorf("failed to list all rooms: %w", err)
		}
		return nil
	case CMD_SEND:
		if c.User.CurrentRoom == nil {
			return fmt.Errorf("failed to send message: user must be in the room")
		}
		message := Message{
			Nickname: c.User.Nickname,
			Text:     c.Args[0],
		}
		c.User.CurrentRoom.Messages = append(c.User.CurrentRoom.Messages, message)
		for _, userID := range c.User.CurrentRoom.Users.List() {
			if userID == c.User.ID {
				continue
			}
			user, ok := c.User.CurrentRoom.Users.Get(userID)
			if !ok {
				continue
			}
			if _, err := reply(user.Conn, message.String()); err != nil {
				log.Println("failed to send message to user: ", err)
				continue
			}
		}
		return nil
	case CMD_QUIT:
		close(c.User.DoneChan)
		return nil
	}
	return fmt.Errorf("unexpected error: unknown command name %q", c.Name)
}
