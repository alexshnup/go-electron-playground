package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demonstration purposes.
	},
}

func tailLogHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()

	cmd := exec.Command("tail", "-f", "/var/log/messages")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
		return
	}

	cmd.Start()

	reader := bufio.NewReader(stdout)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}

		ws.WriteMessage(websocket.TextMessage, line)
	}
}

func topHandler(w http.ResponseWriter, r *http.Request) {

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()

	time.Sleep(2 * time.Second)

	//LastConnectedSSHConnectionID
	client := connectionPool[LastConnectedSSHConnectionID].Client

	fmt.Println("Connected to SSH server.")
	fmt.Printf("Client: %v\n", client)

	session, err := client.NewSession()
	if err != nil {
		log.Println(err)
		return
	}
	defer session.Close()

	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		log.Println("Failed to get StdoutPipe for SSH session:", err)
		return
	}

	if err := session.Start("top -b -d 1"); err != nil { // Start the top command in batch mode
		log.Println("Failed to start SSH command:", err)
		return
	}

	reader := bufio.NewReader(stdoutPipe)

	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}

		ws.WriteMessage(websocket.TextMessage, line)
		fmt.Printf("%s\n", string(line))
	}
}
