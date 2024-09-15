package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	clients     []net.Conn // Liste de tous les clients connectés
	clientNames = make(map[net.Conn]string)
	archive     string
	TimesIns    string
	mutex       sync.Mutex // Mutex pour protéger l'accès aux ressources partagées
)

func main() {
	port := getPort()

	// Écouter sur le port spécifié
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer listener.Close()

	// fmt.Println("Server started. Waiting for clients...")

	// Accepter les connexions des clients
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		mutex.Lock()
		clients = append(clients, conn)
		mutex.Unlock()
		fmt.Println(clients)
		go handleConnection(conn)
	}
}

func getPort() string {
	port := ":8989"
	if len(os.Args) == 2 {
		port = ":" + os.Args[1]
		fmt.Printf("Listening on port %s\n", port)
	} else {
		fmt.Println("[USAGE]: ./TCPChat $port")
		os.Exit(1)
	}
	return port
}

func handleConnection(c net.Conn) {
	if checkChatRoomCapacity(c) {
		return
	}

	defer c.Close()

	sendWelcomeMessage(c)
	handleClientName(c)

	os.WriteFile("archiveData.txt", []byte((archive)), 0o644)
	sendArchivedData(c)
	fmt.Printf("Client %s connected\n", c.RemoteAddr().String())

	for {
		if !processClientMessage(c) {
			return
		}
	}
}

func checkChatRoomCapacity(c net.Conn) bool {
	if len(clients) > 4 {
		c.Write([]byte("Chat room is full. Try again later.\n"))
		removeClient(c)
		c.Close()
		return true
	}
	return false
}

func sendWelcomeMessage(c net.Conn) {
	message := "Welcome to TCP-Chat!\n"
	c.Write([]byte(message))

	asciiArt, _ := os.ReadFile("ascii.txt")
	asciiArt = append(asciiArt, byte('\n'))
	c.Write(asciiArt)
}

func handleClientName(c net.Conn) {
	NbTry := 0
ReName:
	var err error
	mutex.Lock()
	c.Write([]byte("[ENTER YOUR NAME]: "))
	name, _ := bufio.NewReader(c).ReadString('\n')
	name = strings.TrimSpace(name)

	if name == "" || !isValidName(name) {
		c.Write([]byte("We unfortunately don't allow empty names. Please give a valid name\n"))
		err = fmt.Errorf("error")
	}
	for _, client := range clients {
		if client != c {
			clientName, exists := clientNames[client]
			if clientName == name && exists {
				c.Write([]byte("This name is already in use, please try another one\n"))
				err = fmt.Errorf("error")
			}
		}
	}
	mutex.Unlock()
	if err != nil {
		if NbTry < 2 {
			NbTry++
			goto ReName
		} else {
			removeClient(c)
			c.Close()
			return
		}
	}
	mutex.Lock()
	clientNames[c] = name
	mutex.Unlock()
	broadcastMessage(fmt.Sprintf("\n%s has joined our chat...", name), c, "")
	TimesIns = time.Now().Format("2006-01-02 15:04:05")
	AccesMsg(name, c, TimesIns)
}

func isValidName(name string) bool {
	for _, char := range name {
		if char < 32 {
			return false
		}
	}
	return true
}

func AccesMsg(name string, c net.Conn, TimesIns string) {
	for _, client := range clients {
		if client != c {
			clientName, exists := clientNames[client]
			if clientName != name && exists {
				client.Write([]byte(fmt.Sprintf("\n[%s][%s]:", TimesIns, clientName)))
			}
		}
	}
}

func sendArchivedData(c net.Conn) {
	archiveData, _ := os.ReadFile("archiveData.txt")
	if len(archiveData) > 0 {
		c.Write(archiveData)
	}
}

func processClientMessage(c net.Conn) bool {
	name := clientNames[c]
	TimesIns := time.Now().Format("2006-01-02 15:04:05")
	c.Write([]byte(fmt.Sprintf("[%s][%s]:", TimesIns, name)))

	message, err := bufio.NewReader(c).ReadString('\n')
	if err != nil {
		fmt.Printf("%s disconnected.\n", name)
		handleClientDisconnection(c, name)
		return false
	}
	b := true
	if message == "" || containsInvalidChars(message) {
		message = "invalid input!"
		b = false
	}
	message = strings.TrimSpace(message)
	if b {
		mutex.Lock()
		archive += fmt.Sprintf("[%s][%s]:%s\n", TimesIns, name, message)
		mutex.Unlock()
	}
	broadcastMessage(message, c, TimesIns)
	AccesMsg(name, c, TimesIns)
	return true
}

func containsInvalidChars(message string) bool {
	for _, char := range message {
		if (char < 32 && char != 10) || (strings.TrimSpace(message) == "") {
			return true
		}
	}
	return false
}

func handleClientDisconnection(c net.Conn, name string) {
	removeClient(c)
	broadcastMessage(fmt.Sprintf("\n%s has left our chat...", name), c, "")
	TimesIns = time.Now().Format("2006-01-02 15:04:05")
	AccesMsg(name, c, TimesIns)
}

func broadcastMessage(message string, sender net.Conn, TimesIns string) {
	// Verrouiller avant d'accéder à `clients` et `clientNames`
	mutex.Lock()
	defer mutex.Unlock() // Utilisation de `defer` pour garantir que le mutex est toujours libéré
	for _, client := range clients {
		if TimesIns == "" && client != sender {
			clientName, exists := clientNames[client]
			if exists && clientName != "" {
				client.Write([]byte(message))
			}
		} else if client != sender {
			clientName, exists := clientNames[client]
			if exists && clientName != "" {
				client.Write([]byte(fmt.Sprintf("\n[%s][%s]:%s", TimesIns, clientNames[sender], message)))
			}
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
