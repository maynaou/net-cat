package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

var (
	clients     []net.Conn // Liste de tous les clients connectés
	clientNames = make(map[net.Conn]string)
	archive     string
	data        []byte
)

func main() {
	port := ":8989"
	if len(os.Args) == 2 {
		port = ":" + os.Args[1]
		fmt.Printf("Listening on port %s\n", port)
	}
	// Ecouter sur le port spécifié
	l, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer l.Close()

	fmt.Println("Server started. Waiting for clients...")

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		clients = append(clients, conn)
		go handleConnection(conn)
	}
}

// Fonction pour gérer la connexion d'un client
func handleConnection(c net.Conn) {
	if len(clients) > 4 {
		c.Write([]byte("Chat room is full. Try again later.\n"))
		removeClient(c)
		c.Close()
		return
	}
	defer c.Close()
	message := "Welcome to TCP-Chat!\n"
	_, err := c.Write([]byte(message))
	if err != nil {
		fmt.Println(err)
		return
	}
	res, _ := os.ReadFile("ascii.txt")
	res = append(res, byte('\n'))
	// Envoyer un message de bienvenue au client
	_, err = c.Write(res)
	if err != nil {
		fmt.Println(err)
		return
	}

	msg := "[ENTER YOUR NAME]:"
	c.Write([]byte(msg))
	var timestamp string
	name, _ := bufio.NewReader(c).ReadString('\n')
	name = strings.TrimSpace(name)

	clientNames[c] = name
	fmt.Printf("%s has connected.\n", name)
	for _, client := range clients {
		if client != c {
			clientName, exists := clientNames[client]
			if clientName == name && exists {
				c.Write([]byte("invalid name!\n"))
				c.Close()
				return
			}
		}
	}
	broadcastMessage(fmt.Sprintf("\n %s has joined our chat...", name), c, timestamp, name)
	for _, client := range clients {
		if client != c {
			clientName, exists := clientNames[client]
			if clientName != name && exists {
				timestamp = time.Now().Format("2006-01-02 15:04:05")
				client.Write([]byte(fmt.Sprintf("\n[%s][%s]: ", timestamp, clientName)))
			}
		}
	}
	err = os.WriteFile("archiveData.txt", []byte((archive)), 0o644)
	if err != nil {
		fmt.Println("Erreur lors de l'écriture dans le fichier :", err)
		return
	}

	data, _ = os.ReadFile("archiveData.txt")
	if data != nil {
		c.Write(data)
	}

	fmt.Printf("Client %s connected\n", c.RemoteAddr().String())
	for {

		if name == "" {
			c.Write([]byte("invalid name!\n"))
			c.Close()
			return
		}
		for _, char := range name {
			if char < 32 {
				c.Write([]byte("invalid name!\n"))
				c.Close()
				return
			}
		}
		timestamp = time.Now().Format("2006-01-02 15:04:05")
		c.Write([]byte(fmt.Sprintf("[%s][%s]: ", timestamp, name)))
		// Lire le message du client
		msg := ""
		message, err := bufio.NewReader(c).ReadString('\n')
		for _, char := range message {
			if char < 32 || char > 127 {
				continue
			}
			msg += string(char)
		}
		fmt.Println(msg)
		archive += fmt.Sprintf("[%s][%s]: ", timestamp, name) + msg + "\n"
		if err != nil {
			fmt.Printf("%s disconnected.\n", name)
			removeClient(c)
			timestamp = ""
			broadcastMessage(fmt.Sprintf("\n %s has left our chat...", name), c, timestamp, name)
			for _, client := range clients {
				if client != c {
					clientName, exists := clientNames[client]
					if clientName != name && exists {
						timestamp = time.Now().Format("2006-01-02 15:04:05")
						client.Write([]byte(fmt.Sprintf("\n[%s][%s]: ", timestamp, clientName)))
					}
				}
			}
			return
		}

		// Ajouter l'horodatage et le nom du client au message

		msg = strings.TrimSpace(msg)
		fmt.Printf("Received message: %s from %s\n", message, c.RemoteAddr().String())
		timestamp = time.Now().Format("2006-01-02 15:04:05")
		broadcastMessage(msg, c, timestamp, name)
		for _, client := range clients {
			if client != c {
				clientName, exists := clientNames[client]
				if clientName != name && exists {
					timestamp = time.Now().Format("2006-01-02 15:04:05")
					client.Write([]byte(fmt.Sprintf("\n[%s][%s]: ", timestamp, clientName)))
				}

			}
		}
	}
}

// Diffuser un message à tous les clients sauf l'expéditeur
func broadcastMessage(message string, sender net.Conn, timestamp string, name string) {
	for _, client := range clients {
		if timestamp == "" && client != sender {
			clientName, exists := clientNames[client]
			if exists && clientName != "" {
				client.Write([]byte(message))
			}
		} else if client != sender {
			clientName, exists := clientNames[client]
			if exists && clientName != "" {
				client.Write([]byte(fmt.Sprintf("\n[%s][%s]: %s", timestamp, clientNames[sender], message)))
			}
		}
	}
}

func removeClient(c net.Conn) {
	for i, client := range clients {
		if client == c {
			clients = append(clients[:i], clients[i+1:]...)
			delete(clientNames, c)
			fmt.Println(clients)
			break
		}
	}
}
