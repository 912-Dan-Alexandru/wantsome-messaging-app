package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"wantsome-messaging-app/config"
	"wantsome-messaging-app/internal/utils"
	"wantsome-messaging-app/pkg/models"
)

var shutdown os.Signal = syscall.SIGTERM

func initializeMessages(filePath string) []models.Message {
	messagesFromFile, err := readMessagesFromFile(filePath)
	if err != nil {
		utils.LogError(fmt.Sprintf("Error reading messages: %s", err))
		return []models.Message{}
	}
	return messagesFromFile
}

func RunServer() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Error loading config: %s", err)
		return
	}
	serverAddress := fmt.Sprintf("%s:%d", cfg.Server.URL, cfg.Server.Port)
	// serverAddress := fmt.Sprintf("localhost:8080")

	messages = initializeMessages("./messages.txt")
	roomNames, err := readRoomNamesFromFile("./rooms.txt")
	if err != nil {
		fmt.Println("Error reading room names:", err)
	}

	http.HandleFunc("/", home)
	http.HandleFunc("/ws", handleConnections)

	go handleMsg()

	server := &http.Server{Addr: serverAddress}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		log.Printf("Starting server on %s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			log.Printf("error starting server: %s", err)
			stop <- shutdown
		}
	}()

	signal := <-stop
	log.Printf("Shutting down server ... ")

	m.Lock()
	for conn := range userConnections {
		conn.Close()
		delete(userConnections, conn)
	}
	m.Unlock()

	server.Shutdown(nil)
	if signal == shutdown {
		writeMessagesToFile("./messages.txt", messages)
		writeRoomNamesToFile("./rooms.txt", roomNames)
		os.Exit(1)
	}

}
