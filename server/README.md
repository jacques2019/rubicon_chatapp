# Chatapp Server for Rubicon
Chatapp server in the go programming language for the Rubicon interview question/

# Build and Run
Designed to be paired with the bundled client react app, but can be tested using wscat locally

## Option 1 - Container (Recommended)
 - Use the command: "docker build ." while in the server directory
 - Use the command: "docker run \<container name\> -p 8080:8080"

### Dependencies
 - docker OR podman
## Option 2 - Local
 - Install dependencies using: "go mod download"
 - Run with command: "go run main.go"

### dependencies
 - golang

# Notes
Uses the following environment variables:
 - CHATAPP_HOST => Server host (0.0.0.0 default)
 - CHATAPP_PORT => Server port (8080 default)
