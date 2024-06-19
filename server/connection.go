package server

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"minichat/constant"
	"minichat/conversation"
	"minichat/util"
	"net/http"
	"time"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		//CheckOrigin: func(r *http.Request) bool {
		//	return r.Header.Get("Origin") == ""
		//},
	}
)

func HandleWs(w http.ResponseWriter, r *http.Request) {

	// params := mux.Vars(r)
	// roomNumber := params["room_number"]
	// userName := params["username"]
	// password := params["password"]
	// cmd := params["cmd"]

	query := r.URL.Query()
	roomNumber := query.Get("room_number")
	userName := query.Get("username")
	password := query.Get("password")
	cmd := query.Get("cmd")

	log.Printf("Connection Info: RoomNumber is %s, Cmd is %s, UserName is %s, Password is %s\n", roomNumber, cmd, userName, password)

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if roomNumber == "" || roomNumber == "null" || roomNumber == "undefined" {
		log.Println("roomNumber is error")
		defer func(ws *websocket.Conn) {
			err := ws.Close()
			if err != nil {
				log.Printf("ws close error: %+v", err)
			}
		}(ws)

		msg := conversation.Message{
			UserName:   userName,
			Payload:    constant.JoinFailByRoomEmpty,
			RoomNumber: roomNumber,
			Cmd:        constant.CmdJoin,
		}
		byteData, errJsonMarshal := json.Marshal(msg)
		if errJsonMarshal != nil {
			log.Printf("json marshal error, error is %+v\n", errJsonMarshal)
		}
		if errDataSend := util.SocketSend(ws, byteData); errDataSend != nil {
			log.Printf("write message error: %+v\n", errDataSend)
		}
		return
	}

	if room, ok := conversation.Manager.Rooms[roomNumber]; ok { // check password

		if room.Password != password { // password error
			log.Println("password is error")
			defer func(ws *websocket.Conn) {
				err := ws.Close()
				if err != nil {
					log.Printf("ws close error: %+v", err)
				}
			}(ws)
			msg := conversation.Message{
				UserName:   userName,
				Payload:    constant.JoinFailByPassword,
				RoomNumber: roomNumber,
				Cmd:        constant.CmdJoinPasswordFail,
			}
			byteData, err := json.Marshal(msg)
			if err != nil {
				log.Printf("err: %+v\n", err)
			}
			if errDataSend := util.SocketSend(ws, byteData); errDataSend != nil {
				log.Printf("write message error: %+v\n", errDataSend)
			}
			return
		}

		// check name repeat
		for client, _ := range conversation.Manager.Rooms[roomNumber].Clients {
			if client.UserName == userName {
				randomStr := util.RandomString(10)
				userName = userName + randomStr
			}
		}
	}

	// timeout
	timeoutDuration := 1440 * time.Minute
	err = ws.SetReadDeadline(time.Now().Add(timeoutDuration))
	if err != nil {
		log.Printf("SetReadDeadline error: %+v\n", err)
		return
	}

	// heart
	go func() {
		for {
			time.Sleep(30 * time.Second)
			if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second)); err != nil {
				//if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
			log.Println("websocket heartbeat!")
		}
	}()

	// register new client
	client := &conversation.Client{RoomNumber: roomNumber, UserName: userName, Password: password, Conn: ws, Send: make(chan conversation.Message)}

	go client.Write()
	go client.Read()

	conversation.Manager.Register <- client

}
