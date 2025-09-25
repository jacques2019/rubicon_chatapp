package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"os"
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
	ID         int
	Ch         chan string
}

func main() {
	/* Init the logger */
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Server start")

	/* Bind to the socket and creates a listner with an epoll instance (on linux) */
	listener, err := net.Listen(TYPE, ADDRESS)
	defer listener.Close()

	// Being unable to bind is a fatal error and should exit
	if err != nil {
		slog.Error("Unable to bind to socket, closing", "Error", err)
		os.Exit(0)
	}

	clients := make([]ClientConnection, 0, 3) // Default value of three taken from specificaiton of 3 users

	/* Wait for incoming connections and add them to the slice */

	for {
		current_conn, err := listener.Accept()
		if err != nil {
			slog.Error("Unable to accept connections, closing", "Error", err)
			os.Exit(0)
		}

		/* Get the ID of the new client */
		newID := 0
		if len(clients) != 0 {
			newID = clients[len(clients)-1].ID + 1
		}

		slog.Info("New client connected", "ID", newID)

		// Append a new client to the list and start a worker thread pointing to it
		clients := append(clients, ClientConnection{Connection: current_conn, Name: "", ID: newID, Ch: make(chan string, 100)})

		go WorkerThread(clients, len(clients)-1)
	}
}

func WorkerThread(clients []ClientConnection, index int) {
	client := clients[index]

	defer client.Connection.Close()

	/* Read data from the socket and load it into the channels of each client
	*  Should be performant, since net uses epoll under the hood and goroutines are cheap
	 */
	go func() {
		reader := bufio.NewReader(client.Connection)

		for {
			data, err := reader.ReadString('\n')
			if err != nil {
				slog.Error("Socket read error", "Error", err)
				close(client.Ch)
			}

			// Push data into all of the clients channels but their own
			for i := 0; i < len(clients)-1; i++ {
				if i != index {
					clients[i].Ch <- data
				}
			}
		}
	}()

	/* Write data from the channel to the TCP socket */
	go func() {
		for {
			for data := range client.Ch {
				fmt.Println(data)
			}
		}
	}()
}
