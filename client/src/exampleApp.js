import React, { useState, useEffect, useRef } from "react";
import "./App.css"; // Optional styling

const App = () => {
  const [messages, setMessages] = useState([]);
  const [inputValue, setInputValue] = useState("");
  const [username, setUsername] = useState("");
  const [isConnected, setIsConnected] = useState(false);
  const messagesEndRef = useRef(null);
  const wsRef = useRef(null);

  // Auto-scroll to bottom when new messages arrive
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // WebSocket connection (replace with your WebSocket server URL)
  const connectWebSocket = () => {
    if (!username.trim()) return;

    wsRef.current = new WebSocket("ws://localhost:8080");

    wsRef.current.onopen = () => {
      setIsConnected(true);
      console.log("Connected to chat server");
    };

    wsRef.current.onmessage = (event) => {
      const message = JSON.parse(event.data);
      setMessages((prev) => [...prev, message]);
    };

    wsRef.current.onclose = () => {
      setIsConnected(false);
      console.log("Disconnected from chat server");
    };

    wsRef.current.onerror = (error) => {
      console.error("WebSocket error:", error);
      setIsConnected(false);
    };
  };

  // Send message
  const sendMessage = (e) => {
    e.preventDefault();

    if (!inputValue.trim() || !isConnected) return;

    const message = {
      id: Date.now(),
      username,
      text: inputValue,
      timestamp: new Date().toISOString(),
    };

    // Send to WebSocket server
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
    }

    // Add to local messages (or let server echo back)
    setMessages((prev) => [...prev, message]);
    setInputValue("");
  };

  // Disconnect WebSocket on component unmount
  useEffect(() => {
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  // Login form if not connected
  if (!isConnected) {
    return (
      <div className="login-container">
        <h2>Join Chat</h2>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            connectWebSocket();
          }}
        >
          <input
            type="text"
            placeholder="Enter your username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
          />
          <button type="submit">Connect</button>
        </form>
      </div>
    );
  }

  // Main chat interface
  return (
    <div className="chat-container">
      <div className="chat-header">
        <h2>Chat Room</h2>
        <span className="username">Welcome, {username}!</span>
      </div>

      <div className="messages-container">
        {messages.map((message) => (
          <div
            key={message.id}
            className={`message ${message.username === username ? "own-message" : ""}`}
          >
            <div className="message-header">
              <span className="message-username">{message.username}</span>
              <span className="message-time">
                {new Date(message.timestamp).toLocaleTimeString()}
              </span>
            </div>
            <div className="message-text">{message.text}</div>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      <form className="message-form" onSubmit={sendMessage}>
        <input
          type="text"
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          placeholder="Type your message..."
          className="message-input"
        />
        <button type="submit">Send</button>
      </form>
    </div>
  );
};

export default App;
