package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"wantsome-messaging-app/config"
	"wantsome-messaging-app/internal/utils"
	"wantsome-messaging-app/pkg/models"

	"github.com/gorilla/websocket"
)

var current_room string
var current_user string

func RunClient() {
	if len(os.Args) < 2 {
		log.Fatalf("error loging in, type in your username as well")
	}
	username := os.Args[1]
	initial_menu_message := false
	cfg, err := config.LoadConfig()
	if err != nil {
		// log.Printf("Error loading config: %s", err)
		utils.LogInfo(fmt.Sprintf("Error loading config: %s", err))
		return
	}
	// url := fmt.Sprintf("ws://%s:%d/ws", "localhost", 8080)
	url := fmt.Sprintf("ws://%s:%d/ws", cfg.Server.URL, cfg.Server.Port)
	// randId := rand.Intn(10)
	initial_message := models.Message{UserName: fmt.Sprintf("Client %s", username), Type: "login"}

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatalf("error dialing %s\n", err)
	}
	defer c.Close()

	err = c.WriteJSON(initial_message)
	if err != nil {
		log.Fatalf("error ecountered when loging in to the server, try to run the app again")
	}

	done := make(chan bool)

	reader := bufio.NewReader(os.Stdin)

	// reading server messages
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				// log.Printf("error reading: %s\n", err)
				utils.LogError(fmt.Sprintf("error reading: %s\n", err))
				return
			}
			var msg models.Message
			if err := json.Unmarshal(message, &msg); err != nil {
				// log.Printf("error decoding message: %s\n", err)
				utils.LogError(fmt.Sprintf("error decoding message: %s\n", err))
				continue
			}
			handleMessage(msg, username)
			// fmt.Printf("Got message: %s\n", message)
		}
	}()

	// writing messages to server
	go func() {
		for {
			if !initial_menu_message {
				fmt.Println("Welcome to the Messaging App!")
				fmt.Println("0. Exit")
				fmt.Println("1. View Users")
				fmt.Println("2. View Rooms")
				fmt.Println("3. Create Room")
				fmt.Println("4. Join Room")
				fmt.Println("5. Exit current Room")
				fmt.Println("6. Enter chat with user")
				fmt.Println("7. Send message")
				initial_menu_message = true
				fmt.Print("Enter option: \n")
			}

			option, _ := reader.ReadString('\n')
			option = strings.TrimSpace(option)

			switch option {
			case "0":
				os.Exit(0)
			case "1":
				viewUsers(c, username)
			case "2":
				viewRooms(c, username)
			case "3":
				createRoom(reader, c, username)
			case "4":
				joinRoom(reader)
			case "5":
				current_room = ""
			case "6":
				enterChat(reader)
			case "7":
				sendMessage(reader, c, username)
			default:
				fmt.Println("Invalid option, please try again.")
			}
		}
	}()

	<-done
}

func viewUsers(c *websocket.Conn, username string) {
	message := models.Message{
		UserName: username,
		Type:     "list_users",
	}

	err := c.WriteJSON(message)
	if err != nil {
		log.Printf("Error sending request for user list: %s\n", err)
	}
}

func viewRooms(c *websocket.Conn, username string) {
	message := models.Message{
		UserName: username,
		Type:     "list_rooms",
	}

	err := c.WriteJSON(message)
	if err != nil {
		log.Printf("Error sending request for rooms list: %s\n", err)
	}
}

func createRoom(reader *bufio.Reader, c *websocket.Conn, username string) {
	fmt.Print("Enter room name: ")
	roomName, _ := reader.ReadString('\n')

	roomName = strings.TrimSpace(roomName)

	message := models.Message{
		UserName: username,
		Type:     "create_room",
		Message:  roomName,
	}

	err := c.WriteJSON(message)
	if err != nil {
		log.Printf("Error sending request to create room: %s\n", err)
	}
}

func joinRoom(reader *bufio.Reader) {
	fmt.Print("Enter room name: ")
	roomName, _ := reader.ReadString('\n')

	roomName = strings.TrimSpace(roomName)

	current_room = roomName
	current_user = ""
}

func enterChat(reader *bufio.Reader) {
	fmt.Print("Enter user name: ")
	userName, _ := reader.ReadString('\n')

	userName = strings.TrimSpace(userName)

	current_room = ""
	current_user = userName
}

func sendMessage(reader *bufio.Reader, c *websocket.Conn, username string) {
	fmt.Print("Enter message: ")
	messageText, _ := reader.ReadString('\n')

	messageText = strings.TrimSpace(messageText)

	message := models.Message{
		UserName: username,
		Type:     "message",
		Message:  messageText,
		// TimeStamp: time.Now().Format(time.RFC3339),
	}

	if current_room != "" {
		message.Room = current_room
	} else if current_user != "" {
		message.Recipient = current_user
	} else {
		fmt.Println("No recipient or room set.")
		return
	}

	err := c.WriteJSON(message)
	if err != nil {
		log.Printf("Error sending message: %s\n", err)
	}
}

func handleMessage(msg models.Message, username string) {
	if msg.UserName == username {
		if msg.Type == "list_users" || msg.Type == "list_rooms" {
			fmt.Printf("%s\n", msg.Message)
		}
	}
	if msg.Recipient == username && msg.Type == "message" {
		fmt.Printf("%s: %s\n", msg.UserName, msg.Message)
	}
	if msg.Room != "" && msg.Room == current_room && msg.Type == "message" {
		fmt.Printf("Room: %s, %s: %s\n", msg.Room, msg.UserName, msg.Message)
	}
}
