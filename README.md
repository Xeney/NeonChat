## 🏗 **Полный разбор Go-кода**

### 1. **Основная структура сервера**
```go
package main

import (
	"log"
	"net/http"
	"sync"
	
	"github.com/gorilla/websocket"
)
```
- **`package main`** — точка входа в программу.
- **Импорты**:
  - `log` — для вывода ошибок.
  - `net/http` — HTTP-сервер.
  - `sync` — для синхронизации горутин.
  - `github.com/gorilla/websocket` — библиотека для WebSocket.

---

### 2. **Настройка WebSocket**
```go
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}
```
- **`upgrader`** — преобразует HTTP-соединение в WebSocket.
- **`CheckOrigin`** — разрешает подключение с любых доменов (`true` для разработки, в продакшене нужно ограничить!).

---

### 3. **Структуры данных**
```go
type Message struct {
	Text string `json:"text"`
	User string `json:"user"`
}

var (
	clients   = make(map[*websocket.Conn]bool) // Все подключённые клиенты
	clientsMu sync.Mutex                       // Блокировка для безопасного доступа к clients
)
```
- **`Message`** — структура сообщения (текст + отправитель).
- **`clients`** — мапа для хранения активных соединений.
- **`clientsMu`** — мьютекс, чтобы избежать гонки данных при работе с `clients`.

---

### 4. **Главная функция**
```go
func main() {
	http.HandleFunc("/ws", handleConnections) // WebSocket endpoint
	http.Handle("/", http.FileServer(http.Dir("./static"))) // Статические файлы

	log.Println("Сервер запущен на :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```
- **`/ws`** — endpoint для WebSocket.
- **`/`** — отдаёт файлы из папки `./static` (ваш фронтенд).
- Сервер стартует на порту **8080**.

---

### 5. **Обработчик соединений**
```go
func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка подключения:", err)
		return
	}
	defer conn.Close()

	// Добавляем клиента
	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Ошибка чтения:", err)
			removeClient(conn)
			break
		}

		broadcastMessage(msg)
	}
}
```
- **`upgrader.Upgrade`** — "апгрейдит" HTTP до WebSocket.
- **`defer conn.Close()`** — гарантирует закрытие соединения при выходе.
- **`ReadJSON`** — читает сообщение от клиента в формате JSON.
- **`broadcastMessage`** — рассылает сообщение всем.

---

### 6. **Рассылка сообщений**
```go
func broadcastMessage(msg Message) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		err := client.WriteJSON(msg)
		if err != nil {
			log.Println("Ошибка отправки:", err)
			client.Close()
			delete(clients, client)
		}
	}
}
```
- **`WriteJSON`** — отправляет сообщение в JSON-формате.
- Если ошибка — закрываем соединение и удаляем клиента.

---

### 7. **Удаление клиента**
```go
func removeClient(conn *websocket.Conn) {
	clientsMu.Lock()
	delete(clients, conn)
	clientsMu.Unlock()
}
```
- Удаляет клиента из мапы при отключении.

---

## 🧠 **Как это работает на пальцах**
1. **Подключение клиента**:
   ```mermaid
   sequenceDiagram
   Client->>Server: HTTP Request (Upgrade to WebSocket)
   Server->>Client: HTTP 101 (Switching Protocols)
   Server->>clients: Добавляет новое соединение
   ```

2. **Отправка сообщения**:
   ```mermaid
   sequenceDiagram
   ClientA->>Server: {"text": "Привет", "user": "Вася"}
   Server->>ClientB: {"text": "Привет", "user": "Вася"}
   Server->>ClientC: {"text": "Привет", "user": "Вася"}
   ```

3. **Отключение клиента**:
   - Сервер удаляет его из `clients` и перестаёт рассылать сообщения.

---

## 🔥 **Как запустить**
1. Установите библиотеку WebSocket:
   ```bash
   go get github.com/gorilla/websocket
   ```

2. Создайте структуру проекта:
   ```
   chat/
   ├── static/
   │   └── index.html (ваш фронтенд)
   └── main.go
   ```

3. Запустите сервер:
   ```bash
   go run main.go
   ```

4. Откройте в браузере:  
   `http://localhost:8080`

---

## 💡 **Как можно улучшить**
1. **Авторизация**:
   ```go
   func handleConnections(w http.ResponseWriter, r *http.Request) {
       token := r.URL.Query().Get("token")
       if !isValidToken(token) {
           w.WriteHeader(http.StatusUnauthorized)
           return
       }
       // ...остальной код
   }
   ```

2. **Хранение истории в БД**:
   ```go
   // Используйте PostgreSQL/MySQL/SQLite
   db.Exec("INSERT INTO messages (text, user) VALUES (?, ?)", msg.Text, msg.User)
   ```

3. **Лимит сообщений**:
   ```go
   if len(messages) > 100 {
       messages = messages[1:] // Удаляем самое старое
   }
   ```

4. **Команды** (например, `/clear`):
   ```go
   if strings.HasPrefix(msg.Text, "/") {
       handleCommand(msg.Text)
       return
   }
   ```