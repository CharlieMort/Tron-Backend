package main

import (
	"encoding/json"
	"fmt"
	"log"
)

const gridSize int = 10
const gridWidth int = 500
const gridHeight int = 800

type Packet struct {
	Type string `json:"type"` //Type of packet
	Data string `json:"data"` //The actual msg of the data
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Packet
	register   chan *Client
	unregister chan *Client
	rooms      map[string]*Room
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan Packet),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]*Room),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			fmt.Println("Client Connect ------------------------------------------------")
			h.clients[client] = true
			client.SendClientJSON()
		case client := <-h.unregister:
			fmt.Println("Client Disconnect ---------------------------------------------")
			fmt.Println(client.Name)
		case packet := <-h.broadcast:
			fmt.Println("broadcast")
			fmt.Println(packet)
		}
	}
}

type Room struct {
	clients []*Client
	roomID  string
	grid    [][]string
}

func (h *Hub) CreateRoom() string {
	roomID := GetRandomRoomCode()
	h.rooms[roomID] = &Room{
		roomID:  roomID,
		clients: make([]*Client, 0),
		grid:    GetEmptyMap(),
	}
	return roomID
}

func (h *Hub) JoinRoom(roomID string, client *Client) {
	h.rooms[roomID].clients = append(h.rooms[roomID].clients, client)
	client.RoomCode = roomID
}

func (h *Hub) GetGamesList() []string {
	list := make([]string, 0)
	for roomID := range h.rooms {
		list = append(list, roomID)
	}
	return list
}

func (h *Hub) StartGame(roomID string) {
	h.rooms[roomID].grid = GetEmptyMap()
	for idx, client := range h.rooms[roomID].clients {
		ppos := fmt.Sprintf("%x", idx)
		client.SendPacket(Packet{
			Type: "startGame",
			Data: ppos,
		})
	}

	h.SendGameUpdate(roomID)
}

func (h *Hub) SendGameUpdate(roomID string) {
	gridJson, err := json.Marshal(h.rooms[roomID].grid)
	if err != nil {
		log.Fatal(err)
	}
	for _, client := range h.rooms[roomID].clients {
		client.SendPacket(Packet{
			Type: "gameData",
			Data: string(gridJson),
		})
	}
}

func (h *Hub) SendRoom(roomID string, packet Packet) {
	for _, client := range h.rooms[roomID].clients {
		client.SendPacket(packet)
	}
}

func GetEmptyMap() [][]string {
	grid := make([][]string, 0)
	for y := 0; y < gridHeight/gridSize; y++ {
		grid = append(grid, make([]string, 0))
		for X := 0; X < gridWidth/gridSize; X++ {
			grid[y] = append(grid[y], "5")
		}
	}
	return grid
}
