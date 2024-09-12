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
	name, _ := bufio.NewReader(c).ReadString('\n')
	name = strings.TrimSpace(name)
	clientNames[c] = name
	fmt.Printf("%s has connected.\n", name)
	if name != "" {
		broadcastMessage(fmt.Sprintf("\nLee has joined our chat..."), c)
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		c.Write([]byte(fmt.Sprintf("[%s][%s]: ", timestamp, name)))
	}

	fmt.Printf("Client %s connected\n", c.RemoteAddr().String())
	reader := bufio.NewReader(c)
	for {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		c.Write([]byte(fmt.Sprintf("[%s][%s]: ", timestamp, name)))
		// Lire le message du client
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("%s disconnected.\n", name)
			removeClient(c)
			broadcastMessage(fmt.Sprintf("Lee has left our chat..."), c)
			return
		}
		// Ajouter l'horodatage et le nom du client au message
		message = strings.TrimSpace(message)
		fmt.Printf("Received message: %s from %s\n", message, c.RemoteAddr().String())
		broadcastMessage(message, c)
		timestamp = time.Now().Format("2006-01-02 15:04:05")
		formattedMessage := fmt.Sprintf("[%s][%s]: %s", timestamp, name, strings.TrimSpace(message))
		broadcastMessage(formattedMessage, c)
	}
}

// Diffuser un message à tous les clients sauf l'expéditeur
func broadcastMessage(message string, sender net.Conn) {
	for _, client := range clients {
		if client != sender {
			client.Write([]byte(message + "\n"))
		}
	}
}

func removeClient(c net.Conn) {
	for i, client := range clients {
		if client == c {
			clients = append(clients[:i], clients[i+1:]...)
			delete(clientNames, c)
			break
		}
	}
}
