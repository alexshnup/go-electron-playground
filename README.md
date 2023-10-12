# My Playground with Golang+Electron

## JUST FOR FUN

![go-electron-playground](https://github.com/alexshnup/go-electron-playground/assets/4953963/709c801c-df5b-4368-a389-a981992bac5d)

This is a dark-themed Electron application for executing SSH commands with a Golang backend and Websocket for data transfer between Electron and Golang.

## Features
Connect to a server using SSH
Receive stdout from the Linux "top" command

## Requirements
- Golang
- Electron
- Node.js

## Setup

### Clone this repository:
```bash
git clone https://github.com/alexshnup/go-electron-playground.git
```

### Start:

Start the Golang server:
```bash
cd go-electron-playground
go mod tidy
go build . && ./go-electron-playground
```
in another terminal window:
```bash
cd electron-playground-ui
npm install
npm start
```

## Usage

Enter the SSH server address and port into the input fields.
You can use Password or Private Key authentication.
Your private key or password could be saved in encrypted. (still not completed)

Click the "Connect" button.
The output of the "top" command will be displayed in the text area.

Also you car run any other command in the input field and click "Run" button.

### Future Plans
- Learn more about Electron
- Add support for more SSH commands
- Add the ability to execute multiple commands at once
- Add a GUI for managing SSH connections

