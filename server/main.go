package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
)

/* Networking constants */
const (
	HOST    = "127.0.0.1"
	PORT    = "8080"
	TYPE    = "tcp"
	ADDRESS = HOST + ":" + PORT
)

/* Contains all the information for a client connection */
type ClientConnection struct {
	Connection net.Conn
	Name       string
	Ch         chan string
}

/* For transmission */
type MessageContainer struct {
	MessageData string `json:"message"`
	SenderName  string `json:"sender"`
}

/* For receiving */
type IncomingPacket struct {
	Type string `json:"Type"`
	Data string `json:"Data"`
}

/* Protects map with mutex and keeps track of next index */
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

	// If the client exists, close the connection, channel and remove from the map
	if _, ok := cManager.Clients[clientID]; ok {
		cManager.Clients[clientID].Connection.Close()
		close(cManager.Clients[clientID].Ch)
		delete(cManager.Clients, clientID)
	} else {
		slog.Warn("Attempted to remove a non-existing client")
	}
}

func (cManager *ClientManager) Broadcast(senderID int, message string) {
	client, exists := cManager.getClient(senderID)

	cManager.mutex.Lock()
	defer cManager.mutex.Unlock()

	if !exists {
		slog.Info("Attempt to transmit with non existint client", "senderID", senderID)
		return
	}

	name := client.Name

	// Package with JSON
	messageC := MessageContainer{MessageData: message, SenderName: name}
	data, _ := json.Marshal(messageC)

	// Transmit to all open connections
	for id, client := range cManager.Clients {
		if id != senderID {
			client.Ch <- string(data)
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

	// Bind to the socket and creates a listner with an epoll instance (on linux)
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

		// Append a new client to the list and start a worker thread to handle it
		newID := cManager.AddClient(current_conn)
		slog.Info("New client connected", "ID", newID, "Address", current_conn.LocalAddr())

		go WorkerThread(&cManager, newID)
	}
}

func WorkerThread(cManager *ClientManager, clientID int) {
	client, exists := cManager.getClient(clientID)

	if !exists {
		slog.Warn("Tried to fetch client that does not exist", "ID", clientID)
		return
	}

	// Cleanup handled by RemoveClient
	defer slog.Info("Client disconnected and cleaned up", "ID", clientID)

	/* Read data from the socket and load it into the channels of each client
	*  Should be performant, since net uses epoll under the hood and goroutines are cheap
	 */
	NameReceived := false
	go func() {
		for {
			// Fetch data and checl for closed sockets
			reader := bufio.NewReader(client.Connection)
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

			// Parse JSON
			packet := IncomingPacket{}
			err = json.Unmarshal([]byte(data), &packet)
			if err != nil {
				slog.Warn("Unable to parse JSON", "message", data)
			}

			// Handle packet
			switch packet.Type {
			case "join":
				client.Name = packet.Data
				NameReceived = true
				slog.Info("Client Named", "ID", clientID, "name", client.Name)

			case "message":
				// Shouldn't be able to broadcast before being named
				if !NameReceived {
					slog.Warn("Received a message from client before name", "ID", clientID)
				} else {
					cManager.Broadcast(clientID, packet.Data)
				}

			default:
				slog.Warn("Unknown message type received from client", "type", packet.Type)
			}
		}
	}()

	/* Write data from the channel to the TCP socket
	*  Outside of a goroutine to keep thread active
	 */
	for {
		data, ok := <-cManager.Clients[clientID].Ch

		if !ok {
			break
		}

		// Attempt to write to the connection
		_, err := client.Connection.Write([]byte(data))
		if err != nil {
			slog.Info("Socket write error")
			break
		}
	}
}
