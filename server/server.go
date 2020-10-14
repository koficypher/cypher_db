package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

//Server is a tcp server struct for running the in-moemory databse
type Server struct {
	listener         net.Listener
	quit             chan struct{}
	exited           chan struct{}
	db               memoryDB
	connections      map[int]net.Conn
	connCloseTimeout time.Duration
}

//NewServer returns or inits a new server for DB to run on
func NewServer() *Server {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Failed to start listener", err.Error())
	}

	srv := &Server{
		listener:         lis,
		quit:             make(chan struct{}),
		exited:           make(chan struct{}),
		db:               newDB(),
		connections:      map[int]net.Conn{},
		connCloseTimeout: 10 * time.Second,
	}

	go srv.serve()

	return srv
}

func (s *Server) serve() {
	var id int
	fmt.Println("Starting DB server")
	fmt.Println("Listening for client connections")

	for {
		select {
		case <-s.quit:
			fmt.Println("Shutting down sever")
			err := s.listener.Close()
			if err != nil {
				fmt.Println("Could not close listener", err.Error())
			}

			if len(s.connections) > 0 {
				s.warnConnections(s.connCloseTimeout)
				<-time.After(s.connCloseTimeout)
				s.closeConnections()
			}

			close(s.exited)
			return
		default:
			tcpListener := s.listener.(*net.TCPListener)
			err := tcpListener.SetDeadline(time.Now().Add(2 * time.Second))

			if err != nil {
				fmt.Println("Failed to set listener deadline")
			}

			conn, err := tcpListener.Accept()
			if oppErr, ok := err.(*net.OpError); ok && oppErr.Timeout() {
				continue
			}

			if err != nil {
				fmt.Println("Failed to accept connection", err.Error())
			}
			write(conn, "Welcome To Cypher DB")
			s.connections[id] = conn
			go func(connID int) {
				fmt.Println("Client with id", connID, "joined")
				s.handleConnection(conn)
				delete(s.connections, connID)
				fmt.Println("Client with id", connID, "left")

			}(id)
			id++
		}
	}
}

func write(conn net.Conn, s string) {
	_, err := fmt.Fprintf(conn, "%s\n-> ", s)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		terms := strings.ToLower(strings.TrimSpace(scanner.Text()))
		values := strings.Split(terms, " ")

		switch {
		case len(values) == 3 && values[0] == "set":
			s.db.set(values[1], values[2])
			write(conn, "OK")
		case len(values) == 2 && values[0] == "get":
			val := values[1]
			value, found := s.db.get(values[1])
			if !found {
				write(conn, fmt.Sprintf("Key %s not found", val))
			} else {
				write(conn, value)
			}
		case len(values) == 2 && values[0] == "delete":
			s.db.delete(values[1])
			write(conn, "OK")
		case len(values) == 1 && values[0] == "exit":
			if err := conn.Close(); err != nil {
				fmt.Println("Could not exit connection", err.Error())
			}
		default:
			write(conn, fmt.Sprintf("Unknown command %s", terms))
		}
	}
}

func (s *Server) warnConnections(timeout time.Duration) {
	for _, conn := range s.connections {
		write(conn, fmt.Sprintf("Host wants to shutdown server in %s", timeout.String()))
	}

}

func (s *Server) closeConnections() {
	fmt.Println("Closing all connections")
	for id, conn := range s.connections {
		err := conn.Close()
		if err != nil {
			fmt.Println("Could not close connection with id:", id)
		}
	}
}

//Stop - stops the running server
func (s *Server) Stop() {
	fmt.Println("Stopping the DB server")
	close(s.quit)
	<-s.exited
	fmt.Println("Saving memory records to DB")
	s.db.save()
	fmt.Println("DB server successfully stopped")
}
