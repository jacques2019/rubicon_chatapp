package main

import (
	"bufio"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
)

/* Networking constants */
const (
	HOST    = "127.0.0.1"
	PORT    = "9090"
	TYPE    = "tcp"
	ADDRESS = HOST + ":" + PORT
)

/* Contains all the information for a client connection */
type ClientConnection struct {
	Connection net.Conn
	Name       string
	Ch         chan string
}

type ClientManager struct {
	Clients map[int]ClientConnection
	mutex   sync.RWMutex
	NextID  int
}

func (cManager *ClientManager) AddClient(current_conn net.Conn) int {
	cManager.mutex.Lock()
	defer cManager.mutex.Unlock()

	id := cManager.NextID
	cManager.NextID++

	cManager.Clients[id] = ClientConnection{
		Connection: current_conn,
		Name:       "",
		Ch:         make(chan string, 100),
	}

	return id
}

func (cManager *ClientManager) RemoveClient(clientID int) {
	cManager.mutex.Lock()
	defer cManager.mutex.Unlock()

	if _, ok := cManager.Clients[clientID]; ok {
		cManager.Clients[clientID].Connection.Close()
		close(cManager.Clients[clientID].Ch)
		delete(cManager.Clients, clientID)
	} else {
		slog.Warn("Attempted to remove a non-existing client")
	}
}

func (cManager *ClientManager) Broadcast(senderID int, message string) {
	cManager.mutex.Lock()
	defer cManager.mutex.Unlock()

	for id, client := range cManager.Clients {
		if id != senderID {
			client.Ch <- message
		}
	}

	slog.Info("Message broadcast", "sender", senderID, "message", message)
}

func (cManager *ClientManager) getClient(clientID int) (ClientConnection, bool) {
	cManager.mutex.Lock()
	defer cManager.mutex.Unlock()

	client, ok := cManager.Clients[clientID]

	return client, ok
}

func main() {
	/* Init the logger */
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Server start")

	/* Bind to the socket and creates a listner with an epoll instance (on linux) */
	listener, err := net.Listen(TYPE, ADDRESS)
	// Being unable to bind is a fatal error and should exit
	if err != nil {
		slog.Error("Unable to bind to socket, closing", "Error", err)
		os.Exit(0)
	}

	defer listener.Close()

	/* Wait for incoming connections and add them to the slice */
	cManager := ClientManager{Clients: make(map[int]ClientConnection), NextID: 0}
	for {
		current_conn, err := listener.Accept()
		if err != nil {
			slog.Error("Unable to accept connections, closing", "Error", err)
			os.Exit(0)
		}

		// Append a new client to the list and start a worker thread pointing to it
		newID := cManager.AddClient(current_conn)
		slog.Info("New client connected", "ID", newID, "Address", current_conn.LocalAddr())

		// Launch a worker thread to handle the new connection
		go WorkerThread(&cManager, newID)
	}
}

func WorkerThread(cManager *ClientManager, clientID int) {
	// Attempt to fetch the client
	client, exists := cManager.getClient(clientID)

	if !exists {
		slog.Warn("Tried to fetch client that does not exist", "ID", clientID)
		return
	}

	defer func() {
		slog.Info("Client disconnected and cleaned up", "ID", clientID)
	}()
	/* Read data from the socket and load it into the channels of each client
	*  Should be performant, since net uses epoll under the hood and goroutines are cheap
	 */
	go func() {
		reader := bufio.NewReader(client.Connection)
		for {
			data, err := reader.ReadString('\n')
			if err == io.EOF {
				slog.Info("Client disconnected")
				cManager.RemoveClient(clientID)
				break
			} else if err != nil {
				slog.Error("Socket read error", "Error", err)
				cManager.RemoveClient(clientID)
				break
			}

			cManager.Broadcast(clientID, data)
		}
	}()

	/* Write data from the channel to the TCP socket */

	for {
		data, ok := <-cManager.Clients[clientID].Ch

		if !ok {
			break
		}

		_, err := client.Connection.Write([]byte(data))
		if err != nil {
			slog.Info("Socket write error")
			break
		}
	}
}
