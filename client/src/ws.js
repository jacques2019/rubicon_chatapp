const EventEmitter = require("events");

class WebSocketChatClient extends EventEmitter {
    constructor(host = "localhost", port = 8080) {
        super();
        this.host = host;
        this.port = port;
        this.socket = null;
        this.isConnected = false;
        this.userName = "";
    }

    // Async connect to the server socket and send the join message
    // Emits 'connected'
    async connect(userName) {
        this.userName = userName;

        return new Promise((resolve, reject) => {
            // Convert to WebSocket URL
            const wsUrl = `ws://${this.host}:${this.port}/ws`;
            this.socket = new WebSocket(wsUrl);

            this.socket.onopen = () => {
                console.log(`Connected to ${this.host}:${this.port}`);
                this.isConnected = true;

                // Send join message
                this.send({
                    type: "join",
                    data: userName,
                });

                this.emit("connected");
                resolve();
            };

            this.socket.onmessage = (event) => {
                this.handleData(event.data);
            };

            this.socket.onclose = () => {
                console.log("Connection closed");
                this.isConnected = false;
                this.emit("disconnected");
            };

            this.socket.onerror = (error) => {
                console.error("Socket error:", error);
                this.isConnected = false;
                this.emit("error", error);
                reject(error);
            };
        });
    }

    // Handle incoming data
    handleData(data) {
        const dataStr = data.toString();

        // Split by newlines if multiple messages are sent together
        const messages = dataStr.split("\n").filter((msg) => msg.trim());

        messages.forEach((messageStr) => {
            try {
                const message = JSON.parse(messageStr);
                this.handleMessage(message);
            } catch (error) {
                console.error("Invalid message format:", messageStr);
            }
        });
    }

    // Handle parsed messages
    handleMessage(message) {
        switch (message.type) {
            case "message":
                this.emit("message", {
                    text: message.message,
                    senderName: message.sender,
                });
                break;

            case "userList":
                this.emit("userList", message.message);
                break;

            case "join":
                this.emit("userJoined", message.userName);
                break;

            default:
                console.log("Unknown message type:", message.type);
        }
    }

    // Send any message object
    send(messageData) {
        if (
            this.isConnected &&
            this.socket &&
            this.socket.readyState === WebSocket.OPEN
        ) {
            const jsonMessage = JSON.stringify(messageData);
            this.socket.send(jsonMessage);
        } else {
            console.error("Not connected to server");
        }
    }

    // Send a chat message (convenience method)
    sendMessage(text) {
        this.send({
            Type: "message",
            Data: text,
        });
    }

    disconnect() {
        if (this.socket) {
            this.socket.close();
        }
    }
}

module.exports = { WebSocketChatClient };
export default WebSocketChatClient;
