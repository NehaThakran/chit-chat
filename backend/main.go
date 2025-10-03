package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//The upgrader converts HTTP connections to WebSocket connections. CheckOrigin returning true allows all origins
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	ID        interface{} `bson:"_id,omitempty" json:"id,omitempty"`
	Type      string      `bson:"type" json:"type"`
	Content   string      `bson:"content" json:"content"`
	Sender    string      `bson:"sender" json:"sender,omitempty"`
	Recipient string      `bson:"recipient" json:"recipient,omitempty"`
	Room      string      `bson:"room" json:"room,omitempty"`
	Timestamp time.Time   `bson:"timestamp"`
	Delivered bool        `bson:"delivered" json:"delivered,omitempty"`
}

var clients = make(map[string]*websocket.Conn) // Map to store connected clients (key: username, value: websocket connection object)
var rooms = make(map[string]map[string]*websocket.Conn) // Map to store chat rooms (key: room name, value: map of username to websocket connection object)

// handleConnections upgrades HTTP requests to WebSocket connections and handles them
func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}

	username := r.URL.Query().Get("username")
	room := r.URL.Query().Get("room")
	if(username == "") {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}
	clients[username] = ws // Register the user in the global clients map
	// Register the user in the room
	if(rooms[room] == nil) {
		rooms[room] = make(map[string]*websocket.Conn)
	}
	rooms[room][username] = ws 

	//Deliver Undelivered Messages on Reconnect for this user
	// Run in a separate goroutine to avoid blocking
	go func() {
		filter := bson.M{"recipient": username, "delivered": false}
		cursor, err := MessagesCollection.Find(context.TODO(), filter)
		if err != nil {
			fmt.Println("Error fetching undelivered messages: ", err)
			return
		}
		defer cursor.Close(context.TODO())

		for cursor.Next(context.TODO()) {
			var msg Message
			if err := cursor.Decode(&msg); err != nil {
				fmt.Println("Decode error: ", err)
				continue
			}
			msgJSON, err := json.Marshal(msg)
			if err != nil {
				fmt.Println("Marshal error: ", err)
				continue
			}
			ws.WriteMessage(websocket.TextMessage, msgJSON)

			// Mark message as delivered
			update := bson.M{
				"$set": bson.M{
					"delivered": true,
				},
			}
			MessagesCollection.UpdateOne(context.TODO(), bson.M{"_id": msg.ID}, update)
		}
	}()

	defer func(){
		ws.Close()
		delete(clients, username) // Remove from global clients map
		delete(rooms[room], username) // Clean up when user disconnects
	}()

	for {  // Infinite loop to continuously handle WebSocket messages
		_, msg, err := ws.ReadMessage()
		// ReadMessage returns:
		// 1. messageType (ignored with _)
		// 2. message payload (msg)
		// 3. error if any (err)

		if err != nil {
			fmt.Println("Read error: ", err)
			break
		}

		var incoming Message
		if err:= json.Unmarshal(msg, &incoming); err != nil {
			fmt.Println("Unmarshal error: ", err)
			continue
		}

		response := Message {
			Type: incoming.Type,
			Content: incoming.Content,
			Sender: incoming.Sender,
			Recipient: incoming.Recipient,
			Timestamp: time.Now(),
		}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			fmt.Println("Marshal error: ", err)
			continue
		}

		switch incoming.Type {
			case "typing":
				ws.WriteMessage(websocket.TextMessage, responseJSON)
			case "private":
				incoming.Delivered = false
				fmt.Printf("Saving message from %s: %s\n", incoming.Sender, incoming.Content)
				result, err := MessagesCollection.InsertOne(context.TODO(), incoming)
				if err != nil {
					fmt.Println("Insert error: ", err)
					break
				}
				fmt.Println("Inserted message with ID: ", result.InsertedID)

				recipientConn, ok := clients[incoming.Recipient]
				if !ok {
					ws.WriteMessage(websocket.TextMessage, []byte(incoming.Recipient + " is offline. msg saved!"))
					continue
				}
				if ok {
					recipientConn.WriteMessage(websocket.TextMessage, responseJSON)
					ws.WriteMessage(websocket.TextMessage, []byte("Message delivered to " + incoming.Recipient) ) // Acknowledge sender
					//update the message as delivered in DB
					filter := bson.M{
						"sender": incoming.Sender,
						"recipient": incoming.Recipient,
						"timestamp": incoming.Timestamp,
					}
					update := bson.M{
						"$set": bson.M {
							"delivered": true,
						},
					}
					MessagesCollection.UpdateOne(context.TODO(), filter, update)
				}
			case "message":
				// Proceed to save message to DB
				fmt.Printf("Saving message from %s: %s\n", incoming.Sender, incoming.Content)
				result, err := MessagesCollection.InsertOne(context.TODO(), incoming)
				if err != nil {
					fmt.Println("Insert error: ", err)
					break
				}
				fmt.Println("Inserted message with ID: ", result.InsertedID)

				ws.WriteMessage(websocket.TextMessage, responseJSON) // Echo the message back to sender

				// Broadcast to all users in the room
				for _, clientConn := range rooms[incoming.Room] {
					if clientConn != ws { // Don't send back to sender
						clientConn.WriteMessage(websocket.TextMessage, responseJSON)
					}
				}

			default:
				fmt.Println("Unknown message type")
				continue
		}
		
	}

}

// handleHistory handles HTTP requests to fetch message history
func handleHistory(w http.ResponseWriter, r *http.Request) {
	//enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")


	username := r.URL.Query().Get("username")
	room := r.URL.Query().Get("room")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}
	filter := bson.M{
		// Public messages in the room
		"room": room,
    }
	opts := options.Find().SetSort(bson.M{"timestamp": -1}).SetLimit(20) // Last 20 messages
	
	cursor, err := MessagesCollection.Find(context.TODO(), filter, opts)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var messages []Message
	for cursor.Next(context.TODO()) {
		var msg Message
		cursor.Decode(&msg)
		messages = append(messages, msg)
	}
	json.NewEncoder(w).Encode(messages)

}

func main() {
	InitDB() // Initialize the database connection
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/history", handleHistory)
	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}