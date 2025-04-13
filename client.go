package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

type ClientJSON struct {
	Id       string `json:"id"`
	RoomCode string `json:"roomCode"`
	Name     string `json:"name"`
}

type ClientData struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan Packet
}

type Client struct {
	ClientJSON
	ClientData
}

func (client *Client) SendPacket(packet Packet) {
	client.Send <- packet
}

func (client *Client) GetJSON() string {
	dat, err := json.Marshal(client.ClientJSON)
	if err != nil {
		log.Println("Couldnt Parse Client JSON")
		return ""
	}
	return string(dat)
}

func (client *Client) SendClientJSON() {
	dat := client.GetJSON()
	client.SendPacket(Packet{
		Type: "clientData",
		Data: dat,
	})
}

func (client *Client) ReadPackets() {
	defer func() {
		client.Hub.unregister <- client
		client.Conn.Close()
	}()

	for {
		_, packetJson, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}
		var packet Packet
		err = json.Unmarshal(packetJson, &packet)
		fmt.Println("-------------------------------------------------------------------------")
		if err != nil {
			log.Printf("Error ReadPackets(1) %v", err)
		}
		switch packet.Type {
		case "getGames":
			gameList, err := json.Marshal(client.Hub.GetGamesList())
			if err != nil {
				log.Fatal(err)
			}
			client.SendPacket(Packet{
				Type: "gameList",
				Data: string(gameList),
			})
		case "hostGame":
			roomID := client.Hub.CreateRoom()
			client.Hub.JoinRoom(roomID, client)
		case "joinGame":
			roomID := packet.Data
			client.Hub.JoinRoom(roomID, client)
			client.Hub.StartGame(roomID)
		case "gameUpdate":
			pos := strings.Split(packet.Data, ",")
			x, _ := strconv.Atoi(pos[0])
			y, _ := strconv.Atoi((pos[1]))
			client.Hub.rooms[client.RoomCode].grid[y][x] = pos[2]
			client.Hub.SendRoom(client.RoomCode, Packet{
				Type: "gameData",
				Data: packet.Data,
			})
		case "playerDead":
			client.Hub.EndRound(client.RoomCode, client)
		}
	}
}

func (client *Client) WritePackets() {
	defer func() {
		log.Panicln("Write Packet Close")
		client.Conn.Close()
	}()

	for {
		select {
		case packet, ok := <-client.Send:
			if !ok {
				log.Println("Error WritePackets (0)")
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			var err error
			err = client.Conn.WriteJSON(packet)
			if err != nil {
				log.Println("Error WritePackets (1)")
				log.Println(err)
			}
		}
	}
}
