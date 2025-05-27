package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Message struct {
	Text string `json:"text"`
	User string `json:"user"`
}

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

func main() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}
		defer conn.Close()

		// Добавляем нового клиента
		clientsMu.Lock()
		clients[conn] = true
		clientsMu.Unlock()

		for {
			var msg Message
			err := conn.ReadJSON(&msg)
			if err != nil {
				log.Println("Read error:", err)
				break
			}

			// Рассылаем сообщение всем клиентам
			broadcast(msg)
		}

		// Удаляем клиента при отключении
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
	})

	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}

func broadcast(msg Message) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		err := client.WriteJSON(msg)
		if err != nil {
			log.Println("Write error:", err)
			client.Close()
			delete(clients, client)
		}
	}
}
