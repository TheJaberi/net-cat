package penguin

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

func HandleClient(connection net.Conn) {
	// Read welcome message from welcome.txt from the message folder
	welcomeMsg, err := readWelcomeMsg()
	if err != nil {
		log.Printf("Error reading welcome.txt: %s", err.Error())

		return
	}

	// Write the welcome message to the new client
	writer := bufio.NewWriter(connection)
	_, err = writer.WriteString(welcomeMsg)
	if err != nil {
		log.Printf("Error sending welcome message to %s: %s", connection.RemoteAddr().String(), err.Error())
		connection.Close()
		UserCounter--
		return
	}
	writer.Flush()

	// Receive client's name
	reader := bufio.NewReader(connection)
	clientName, err := reader.ReadString('\n')
	if err != nil {
		connection.Close()
		UserCounter--
		return
	}
	// trim spaces from client's name
	clientName = strings.TrimSpace(clientName)
	if clientName == "" {
		writer.WriteString("Name cannot be empty. Reconnect\n")
		writer.Flush()
		connection.Close()
		UserCounter--
		return
	}
	for i := 0; i < 1; i++ {
		for _, client := range Clients {
			if clientName == client.Name {
				clientName = clientName + "2"
				i = -1
			}
		}
	}
	// reader2 := bufio.NewReader(connection)
	connection.Write([]byte("Choose Group(1-adnan, 2-abdeen or 3-alali):"))
	choosen, err := reader.ReadString('\n')
	choosen = strings.ReplaceAll(choosen, "\n", "")
	fmt.Print(choosen)
	if err != nil {
		connection.Close()
		UserCounter--
		return
	}
	var choosengroup string
	switch {
	case choosen == "1" || choosen == "adnan":
		choosengroup = "adnan"
	case choosen == "2" || choosen == "abdeen":
		choosengroup = "abdeen"
	case choosen == "3" || choosen == "alali":
		choosengroup = "alali"
default:
	choosengroup = "adnan" 
	connection.Write([]byte("default group chat adnan choosen\n"))
	}
		// will show chat history for users that join later
	if len(AllMessages) != 0 {
		connection.Write([]byte("\n----------------------history----------------------\n"))
	}
	for _, pastMessage := range AllMessages {
		connection.Write([]byte(pastMessage))
	}
	if len(AllMessages) != 0 {
		connection.Write([]byte("----------------------history----------------------\n"))
	}
fmt.Println(choosengroup)
	// Create a Client struct and add it to the clients map

	currentClient := Client{Name: clientName, Socket: connection, Group: choosengroup}
	Clients[connection] = currentClient

	// announce to all clients, the name of who joined our chat
	for _, client := range Clients {
		if currentClient.Socket != client.Socket && currentClient.Group == client.Group{
			client.Socket.Write([]byte("\n" + currentClient.Name + " has joined "+ currentClient.Group+" group chat...\n"))
			// client.Socket.W
			client.Socket.Write([]byte("[" + time.Now().Format("2006-01-02 15:04:05") + "][" + client.Name + "]: "))
		}
	}
	AllMessages = append(AllMessages,currentClient.Name + " has joined "+ currentClient.Group+" group chat...\n")

	// go routine that will keep reading each clients input
	go func() {
		defer connection.Close() // after programming is done running, it will make sure to close connection

		contreader := bufio.NewReader(connection) // variable of type reader(has capability to read)
		connection.Write([]byte("[" + time.Now().Format("2006-01-02 15:04:05") + "][" + currentClient.Name + "]: "))
		for {
			clientMessage, err := contreader.ReadString('\n') // reads everything until first occurence of new line
			if err != nil {                                   // anytime an error happens, assume user has disconnected. errors could be EOF which means they did a signal interrupt
				for _, client := range Clients { // broadcast message to all users that current client disconnected
					if currentClient.Socket != client.Socket && currentClient.Group == client.Group { // send to all clients that someone left, except that person
						client.Socket.Write([]byte("\n" + currentClient.Name + " has left our chat...\n"))
						client.Socket.Write([]byte("[" + time.Now().Format("2006-01-02 15:04:05") + "][" + client.Name + "]: "))
					}
				}
				AllMessages = append(AllMessages, currentClient.Name + " has left "+ currentClient.Group+" group chat...\n")
				connection.Close()
				UserCounter--
				fmt.Println(UserCounter)
				return
			}
			if len(clientMessage) > 1 && clientMessage[0:2] == "--" { // check for flag
				Flags(clientMessage, connection, currentClient)
				currentClient = Clients[connection]
			} else {
				// will check if client tries sending an empty message, if so it won't broadcast it
				clientMessage = strings.TrimSpace(clientMessage)
				if clientMessage == "" {
					connection.Write([]byte("[" + time.Now().Format("2006-01-02 15:04:05") + "][" + currentClient.Name + "]: "))
					continue
				}
				// fmt.Print("[" + time.Now().Format("2006-01-02 15:04:05") + "][" + currentClient.Name + "]: " + clientMessage) // XXX

				// append to all messages slice, which stores all messages
				AllMessages = append(AllMessages, "["+time.Now().Format("2006-01-02 15:04:05")+"]["+currentClient.Name+"]: "+clientMessage+"\n")

				// where messages are sent to be printed
				currentClient.Message = clientMessage
				messages <- currentClient // channel to communicate with broadcast message go routine, sends data of type client, along with his socket, message and name
			}
		}
	}()
}
