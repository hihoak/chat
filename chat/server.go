package chat

import (
	"fmt"
	"log"
	"net"
	"sync"
)

const (
	defaultAddr = "localhost"
	port        = "8080"
)

type Users struct {
	mu   *sync.RWMutex
	data map[int]*User
}

func (u *Users) Get(key int) (*User, bool) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	user, ok := u.data[key]
	return user, ok
}

func (u *Users) Set(key int, value *User) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.data[key] = value
}

func (u *Users) List() []int {
	res := make([]int, 0, len(u.data))
	u.mu.RLock()
	for key := range u.data {
		res = append(res, key)
	}
	u.mu.RUnlock()
	return res
}

func (u *Users) Delete(key int) {
	u.mu.Lock()
	defer u.mu.Unlock()
	delete(u.data, key)
}

type Server struct {
	Listener net.Listener
	Users    Users
	Rooms    Rooms
	Commands chan Command
}

func NewServer() (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", defaultAddr, port))
	if err != nil {
		return nil, fmt.Errorf("failed to start listen on: :%s: %w", port, err)
	}
	return &Server{
		Listener: listener,
		Users: Users{
			mu:   &sync.RWMutex{},
			data: make(map[int]*User),
		},
		Rooms: Rooms{
			mu:   &sync.RWMutex{},
			data: make(map[string]*Room),
		},
	}, nil
}

func (s *Server) Run() error {
	defer s.Listener.Close()
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			log.Println("ERROR: failed to accept connection: ", err)
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	user := NewUser(conn, s)
	s.Users.Set(user.ID, user)
	defer s.Users.Delete(user.ID)

	log.Println("Starting a new user session")
	if err := user.StartSession(); err != nil {
		log.Println("ERROR: got some error in user session: ", err)
	}
	log.Println("end session")
}
