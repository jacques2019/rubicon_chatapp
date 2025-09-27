package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

/* Networking constants */
const (
	HOST    = "127.0.0.1"
	PORT    = "8080"
	ADDRESS = HOST + ":" + PORT
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

/* Contains all the information for a client connection */
type ClientConnection struct {
	Connection *websocket.Conn
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

func (cManager *ClientManager) AddClient(current_conn *websocket.Conn) int {
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

	// Package with JSON
	messageC := MessageContainer{MessageData: message, SenderName: client.Name}
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

func handleWebSocket(cManager *ClientManager, w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Failed to upgrade to WebSocket", "Error", err)
		return
	}

	// Append a new client to the list and start a worker thread to handle it
	newID := cManager.AddClient(conn)
	slog.Info("New client connected", "ID", newID, "Address", conn.LocalAddr())

	go WorkerThread(cManager, newID)
}

func main() {
	/* Init the logger */
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Server start")

	/* Wait for incoming connections and add them to the slice */
	cManager := ClientManager{Clients: make(map[int]ClientConnection), NextID: 0}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(&cManager, w, r)
	})

	slog.Info("WebSocket server listening", "address", ADDRESS)
	err := http.ListenAndServe(ADDRESS, nil)
	if err != nil {
		slog.Error("Unable to start server", "Error", err)
		os.Exit(0)
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
			// Fetch data and check for closed sockets
			_, data, err := client.Connection.ReadMessage()
			if err != nil {
				slog.Info("Client disconnected", "Error", err)
				cManager.RemoveClient(clientID)
				break
			}

			// Parse JSON
			packet := IncomingPacket{}
			err = json.Unmarshal(data, &packet)
			if err != nil {
				slog.Warn("Unable to parse JSON", "message", string(data))
			}

			// Handle packet
			switch packet.Type {
			case "join":
				client.Name = string(packet.Data)

				// Update the map client to the modified struct
				cManager.mutex.Lock()
				cManager.Clients[clientID] = client
				cManager.mutex.Unlock()

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

	/* Write data from the channel to the WebSocket connection
	*  Outside of a goroutine to keep thread active
	 */
	for {
		data, ok := <-cManager.Clients[clientID].Ch

		if !ok {
			break
		}

		// Attempt to write to the connection
		err := client.Connection.WriteMessage(websocket.TextMessage, []byte(data))
		if err != nil {
			slog.Info("Socket write error", "Error", err)
			break
		}
	}
}
