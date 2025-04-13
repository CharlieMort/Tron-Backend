package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const PORT = ":3001"
const DEBUG = false

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
}

func GetRandomRoomCode() string {
	rnd := "0123456789"
	lng := 5
	out := ""
	for i := 0; i < lng; i++ {
		out = out + string(rnd[rand.Intn(len(rnd))])
	}
	return out
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("They Joined")
	client := &Client{
		ClientJSON{
			Id:       strconv.Itoa(nextClientID),
			RoomCode: "",
			Name:     "",
		},
		ClientData{
			Hub:  hub,
			Conn: conn,
			Send: make(chan Packet),
		},
	}
	nextClientID += 1
	client.Hub.register <- client

	go client.WritePackets()
	go client.ReadPackets()
}

func Chk(r *http.Request) bool {
	return true
}

var nextClientID int
var upgrader = websocket.Upgrader{
	CheckOrigin: Chk,
}

func main() {
	fmt.Println("Tron Backend Server Running...")

	nextClientID = 1
	hub := NewHub()
	go hub.run()

	r := mux.NewRouter()
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		log.Println("Someone Connected to ws")
		serveWs(hub, w, r)
	})

	http.ListenAndServe(PORT, r)
}
