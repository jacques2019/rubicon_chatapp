# Basic Chatapp
Basic chatapp for the Rubicon interview task.

# Caveats
 - The client can only be run on the same machine it's hosted on, as it points to localhost for the server websocket.
 - Only tested on linux.

# Installation
## Dependencies
Relies on the system having docker and docker compose (or podman) installed.

## Run
 - Navigate to the parent directory (chatapp)
 - Build the images with: "docker compose build"
 - Start the containers with: "docker compose up -d"
 - NOTE: run without "-d" to see log and error messages
