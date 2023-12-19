package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"wantsome-messaging-app/internal/utils"
	"wantsome-messaging-app/pkg/models"

	"github.com/gorilla/websocket"
)

var messages []models.Message

var (
	m               sync.Mutex
	userConnections = make(map[*websocket.Conn]string)
	broadcast       = make(chan models.Message)
	// rooms           = make(map[string]*Room)
	roomNames = []string{}
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func readMessagesFromFile(filePath string) ([]models.Message, error) {
	var messages []models.Message

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return messages, err
	}

	err = json.Unmarshal(data, &messages)
	if err != nil {
		return messages, err
	}

	return messages, nil
}

func writeMessagesToFile(filePath string, messages []models.Message) error {
	data, err := json.Marshal(messages)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func readRoomNamesFromFile(filePath string) ([]string, error) {
	var roomNames []string

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return roomNames, err
	}

	roomNames = strings.Split(string(data), "\n")
	return roomNames, nil
}

// func writeRoomNamesToFile(roomNames []string) error {
// 	file, err := os.Create("./rooms.txt")
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	writer := bufio.NewWriter(file)
// 	for _, name := range roomNames {
// 		_, err := fmt.Fprintln(writer, name)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return writer.Flush()
// }

func writeRoomNamesToFile(filePath string, roomNames []string) error {
	data := []byte(strings.Join(roomNames, "\n"))

	err := ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func home(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprintf(w, "Hello world from my server!")
	utils.LogInfo("Hello world from my server!")
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// fmt.Printf("got error upgrading connection %s\n", err)
		utils.LogError(fmt.Sprintf("got error upgrading connection: %s", err))
		return
	}
	defer conn.Close()

	m.Lock()
	userConnections[conn] = ""
	m.Unlock()
	// fmt.Printf("connected client!")
	utils.LogInfo("connected client!")
	// roomNames, err = readRoomNamesFromFile("./rooms.txt")

	for {
		var msg models.Message = models.Message{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			// fmt.Printf("got error reading message %s\n", err)
			utils.LogError(fmt.Sprintf("got error reading message %s\n", err))
			m.Lock()
			delete(userConnections, conn)
			m.Unlock()
			return
		}
		m.Lock()
		if msg.Type == "login" {
			userConnections[conn] = msg.UserName
		}
		// broadcast <- msg
		m.Unlock()
		if msg.Type == "list_users" {
			userList := make([]string, 0, len(userConnections))
			for _, username := range userConnections {
				userList = append(userList, username)
			}
			usersMessage := strings.Join(userList, "; ")
			msg_send := models.Message{
				Type:     "list_users",
				Message:  usersMessage,
				UserName: msg.UserName,
			}
			// conn.WriteJSON(msg_send)
			broadcast <- msg_send
		}
		if msg.Type == "list_rooms" {
			roomList := make([]string, 0, len(roomNames))
			for _, roomName := range roomNames {
				roomList = append(roomList, roomName)
			}
			roomsMessage := strings.Join(roomList, "; ")
			msg_send := models.Message{
				Type:     "list_rooms",
				Message:  roomsMessage,
				UserName: msg.UserName,
			}
			// conn.WriteJSON(msg_send)
			broadcast <- msg_send
		}
		if msg.Type == "create_room" {
			roomNames = append(roomNames, msg.Message)
		}
		if msg.Type != "create_room" && msg.Type != "list_rooms" && msg.Type != "list_users" && msg.Type != "login" {
			messages = append(messages, msg)
			m.Lock()
			userConnections[conn] = msg.UserName
			m.Unlock()
			broadcast <- msg
		}
	}
}

func handleMsg() {
	for {
		msg := <-broadcast

		m.Lock()
		for client, username := range userConnections {
			if username != msg.UserName {
				err := client.WriteJSON(msg)
				if err != nil {
					// fmt.Printf("got error broadcating message to client %s", err)
					utils.LogError(fmt.Sprintf("got error broadcating message to client %s", err))
					client.Close()
					delete(userConnections, client)
				}
			}
		}
		m.Unlock()
	}
}
