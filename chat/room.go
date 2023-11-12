package chat

import (
	"fmt"
	"strings"
	"sync"
)

type Rooms struct {
	mu   *sync.RWMutex
	data map[string]*Room
}

func (r *Rooms) ListNames() []string {
	r.mu.RLock()
	ans := make([]string, 0, len(r.data))
	for _, room := range r.data {
		ans = append(ans, room.Name)
	}
	r.mu.RUnlock()
	return ans
}

func (r *Rooms) Create(name string) (*Room, error) {
	r.mu.RLock()
	_, ok := r.data[name]
	r.mu.RUnlock()
	if ok {
		return nil, fmt.Errorf("room with name %q already exists", name)
	}
	room := &Room{
		Name:     name,
		Messages: nil,
		Users: Users{
			mu:   &sync.RWMutex{},
			data: make(map[int]*User),
		},
	}
	r.mu.Lock()
	r.data[name] = room
	r.mu.Unlock()
	return room, nil
}

func (r *Rooms) Get(name string) (*Room, bool) {
	r.mu.RLock()
	room, ok := r.data[name]
	r.mu.RUnlock()
	return room, ok
}

type Room struct {
	Name     string
	Messages []Message
	Users    Users
}

func (r *Room) FormatMessages() string {
	res := strings.Builder{}
	for _, message := range r.Messages {
		res.WriteString(fmt.Sprintf("%s\n", message.String()))
	}
	return res.String()
}

type Message struct {
	Nickname string
	Text     string
}

func (m *Message) String() string {
	return fmt.Sprintf("%s: %s", m.Nickname, m.Text)
}
