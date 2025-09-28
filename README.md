# Basic Chatapp
Basic chatapp for the Rubicon interview task.

# Caveats
 - The client can only be run on the same machine it's hosted on, as it points to localhost for the server websocket.
 - Only tested on linux.

# Run
Relies on the system having docker and docker compose (or podman) installed.

 - Navigate to the parent directory (chatapp)
 - Build the images with: "docker compose build"
 - Start the containers with: "docker compose up -d"
 - NOTE: run without "-d" to see log and error messages

# Usage
 - Using a browser on the host machine open http://127.0.0.1:9090
 - On load, the page will show a login screen. Enter a username and press login or hit "enter"
 - Use the text input bar at the bottom of the screen to send messages using either the send icon or hitting "enter"
 - Active users are shown in the bar on the left (automatically updated)
 - Your connection status can bee seen below the Header and welcome message

# Installation
If you wish for the app to run on boot, set the docker daemon to run on boot and to start the containers on daemon start:
 - Change the docker compose to reflect "restart: always" for both services
 - Ensure the docker daemon runs on boot with "systemctl enable docker" (may need to run as root)