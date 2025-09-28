# Chatapp Client for Rubicon
Chatapp client in the javascript programming language, using the react framework for the Rubicon interview question

# Build and Run
Designed to be paired with the bundled server app. Uses the parcel package for a simple, no-config build system. See package.json to see how the run commands are configured.

## Option 1 - Container (Recommended)
 - Use the command: "docker build ." while in the server directory
 - Use the command: "docker run \<container name\> -p 8090:80"

### Dependencies
 - docker OR podman

## Option 2 - Local
 - Install dependencies using: "npm install"
 - Run with command: "npm start"
 - Build using the command "npm run build"
### dependencies
 - nodejs + npm