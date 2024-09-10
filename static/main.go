package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Score struct {
	Name  string `json:"name"`
	Rank  int    `json:"rank"`
	Score int    `json:"score"`
	Time  int    `json:"time"`
}

var (
	scoreData  Score
	allScores  []Score
	clients    = make(map[*websocket.Conn]bool)
	clientsMux sync.Mutex
	upgrader   = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/story", story)
	http.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")

		http.ServeFile(w, r, "index.html")
	})

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/receive", handler)
	http.HandleFunc("/ws", handleWebSocket)

	http.ListenAndServe("0.0.0.0:8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")

	tmpl, err := template.ParseFiles("home.html")
	if err != nil {
		fmt.Println(err)
		return
	}
	tmpl.Execute(w, allScores)
}

func story(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")

	tmpl, err := template.ParseFiles("story.html")
	if err != nil {
		fmt.Println(err)
		return
	}
	tmpl.Execute(w, r)
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")

	fmt.Println(r.Body)
	err := json.NewDecoder(r.Body).Decode(&scoreData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("Received variable:", scoreData)

	allScores = append(allScores, scoreData)
	manualSortAndUpdateRanks()
	broadcastScores()
}

func manualSortAndUpdateRanks() {
	for i := 0; i < len(allScores); i++ {
		maxIndex := i
		for j := i + 1; j < len(allScores); j++ {
			if allScores[j].Score > allScores[maxIndex].Score {
				maxIndex = j
			}
		}
		allScores[i], allScores[maxIndex] = allScores[maxIndex], allScores[i]
	}

	for i := range allScores {
		allScores[i].Rank = i + 1
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error while upgrading connection:", err)
		return
	}
	defer conn.Close()

	clientsMux.Lock()
	clients[conn] = true
	clientsMux.Unlock()

	defer func() {
		clientsMux.Lock()
		delete(clients, conn)
		clientsMux.Unlock()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func broadcastScores() {
	clientsMux.Lock()
	defer clientsMux.Unlock()

	for client := range clients {
		err := client.WriteJSON(allScores)
		if err != nil {
			fmt.Println("Error sending message:", err)
			client.Close()
			delete(clients, client)
		}
	}
}
