import React, { useState, useEffect, useRef } from "react";
import { Send, User, Users } from "lucide-react";
import { WebSocketChatClient } from "./ws.js";
// Setup the persistent TCP client

function App() {
    const ip = "127.0.0.1";
    const port = 9090;

    const [messages, setMessages] = useState([]);
    const [activeUsers, setActiveUsers] = useState([]);
    const [inputText, setInputText] = useState("");
    const [userName, setUserName] = useState("");
    const [nameInput, setNameInput] = useState("");
    const [hasEnteredName, setHasEnteredName] = useState(false);
    const [isConnected, setIsConnected] = useState(false);

    // Setup the persistent WebSocket client
    const wsClientRef = useRef(null);

    // Set up WS connection when user enters name
    useEffect(() => {
        if (hasEnteredName && !wsClientRef.current) {
            initializeWSConnection();
        }

        // Cleanup on unmount
        return () => {
            if (wsClientRef.current) {
                wsClientRef.current.disconnect();
            }
        };
    }, [hasEnteredName, userName]);

    const initializeWSConnection = async () => {
        try {
            // Create new WS client instance
            wsClientRef.current = new WebSocketChatClient(ip, port);

            setupEventListeners();

            // Connect to server
            await wsClientRef.current.connect(userName);
        } catch (error) {
            console.error("Failed to connect:", error);
            setIsConnected(false);
        }
    };

    const setupEventListeners = () => {
        const client = wsClientRef.current;

        // Listen for incoming messages
        client.on("message", (messageData) => {
            console.log("Received message:", messageData);
            // Add message to React state
            const newMessage = {
                text: messageData.text,
                id: `${messageData.senderName}-${Date.now()}-${Math.random()}`,
                senderName: messageData.senderName,
            };

            console.log("message sender", messageData.sender);
            setMessages((prevMessages) => [...prevMessages, newMessage]);
        });

        client.on("userList", (users) => {
            console.log("Incoming user list", users);

            const formattedUsers = users.map((userName) => ({
                id: userName,
                name: userName,
                avatar: userName.substring(0, 2).toUpperCase(),
            }));

            console.log("Formatted users:", formattedUsers);
            setActiveUsers(formattedUsers);
        });

        client.on("connected", () => {
            console.log("Connected to server");
            setIsConnected(true);
        });

        client.on("error", (error) => {
            console.log("Server connection error", error);
        });
    };

    // Create message structure and send using WS client
    const handleSendMessage = () => {
        if (inputText.trim() !== "" && hasEnteredName) {
            wsClientRef.current.sendMessage(inputText);
            setInputText("");
        }
    };

    const handleKeyPress = (e) => {
        if (e.key === "Enter") {
            handleSendMessage();
        }
    };

    const handleNameSubmit = () => {
        if (nameInput.trim() !== "") {
            setUserName(nameInput.trim());
            setHasEnteredName(true);
        }
    };

    const handleNameKeyPress = (e) => {
        if (e.key === "Enter") {
            handleNameSubmit();
        }
    };

    // Show name entry screen if user hasn't entered their name
    if (!hasEnteredName) {
        return (
            <div className="flex items-center justify-center h-screen bg-gray-900">
                <div className="bg-gray-800 p-8 rounded-lg shadow-lg max-w-md w-full mx-4">
                    <div className="text-center mb-6">
                        <User className="w-16 h-16 text-purple-400 mx-auto mb-4" />
                        <h1 className="text-2xl font-bold text-gray-100 mb-2">
                            Welcome to Chat App
                        </h1>
                        <p className="text-gray-400">Enter your name to start chatting</p>
                    </div>

                    <div className="space-y-4">
                        <input
                            type="text"
                            value={nameInput}
                            onChange={(e) => setNameInput(e.target.value)}
                            onKeyPress={handleNameKeyPress}
                            placeholder="Enter your name..."
                            className="w-full border border-gray-600 bg-gray-900 text-gray-100 placeholder-gray-400 rounded-lg px-4 py-3 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                            autoFocus
                        />
                        <button
                            onClick={handleNameSubmit}
                            disabled={nameInput.trim() === ""}
                            className="w-full bg-purple-500 hover:bg-purple-600 disabled:bg-gray-600 disabled:cursor-not-allowed text-white py-3 rounded-lg transition-colors duration-200 font-medium"
                        >
                            Join Chat
                        </button>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="flex h-screen bg-gray-900">
            {/* Active Users Sidebar */}
            <div className="w-64 bg-gray-800 border-r border-gray-700 flex flex-col">
                <div className="bg-gray-750 p-4 border-b border-gray-700">
                    <div className="flex items-center space-x-2">
                        <Users className="w-5 h-5 text-purple-300" />
                        <h2 className="font-semibold text-gray-100">Active Users</h2>
                        <span className="bg-purple-500 text-white text-xs px-2 py-1 rounded-full">
                            {activeUsers.length}
                        </span>
                    </div>
                </div>

                <div className="flex-1 overflow-y-auto">
                    {activeUsers.map((user) => (
                        <div
                            key={user.id}
                            className="flex items-center space-x-3 p-3 hover:bg-gray-700 cursor-pointer border-b border-gray-700"
                        >
                            <div className="w-10 h-10 bg-purple-500 rounded-full flex items-center justify-center text-white font-medium text-sm">
                                {user.avatar}
                            </div>
                            <div className="flex-1 min-w-0">
                                <p className="text-sm font-medium text-gray-100 truncate">
                                    {user.name}
                                </p>
                            </div>
                        </div>
                    ))}
                </div>
            </div>

            {/* Main Chat Area */}
            <div className="flex-1 flex flex-col">
                {/* Header */}
                <div className="bg-purple-600 text-white p-4 shadow-lg">
                    <div className="flex items-center space-x-3">
                        <User className="w-8 h-8" />
                        <div>
                            <h1 className="text-xl font-semibold">Chat App</h1>
                            <p className="text-purple-100 text sm">{isConnected ? `Connected` : `Disconnected`}</p>
                            <p className="text-purple-100 text-sm">Welcome, {userName}</p>
                        </div>
                    </div>
                </div>

                {/* Messages Container */}
                <div className="flex-1 overflow-y-auto p-4 space-y-4 bg-gray-900">
                    {messages.map((message) => (
                        <div key={message.id}>
                            <div
                                className={`max-w-xs lg:max-w-md px-4 py-2 rounded-lg bg-white-800 text-black-100 shadow-md`}
                            >
                                <p className="text-sm text-purple-400">{message.senderName}</p>
                                <p className="text-sm">{message.text}</p>
                            </div>
                        </div>
                    ))}
                </div>

                {/* Input Area */}
                <div className="bg-gray-800 border-t border-gray-700 p-4">
                    <div className="flex space-x-2">
                        <input
                            type="text"
                            value={inputText}
                            onChange={(e) => setInputText(e.target.value)}
                            onKeyPress={handleKeyPress}
                            placeholder="Type a message..."
                            className="flex-1 border border-gray-600 bg-gray-900 text-gray-100 placeholder-gray-400 rounded-lg px-4 py-2 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                        />
                        <button
                            onClick={handleSendMessage}
                            className="bg-purple-500 hover:bg-purple-600 text-white p-2 rounded-lg transition-colors duration-200"
                        >
                            <Send className="w-5 h-5" />
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default App;
