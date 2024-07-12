package conversation

import (
	"github.com/gorilla/websocket"
	"log"
	"minichat/constant"
	"strings"
	"sync"
)

type ConversationManager struct {
	Rooms          map[string]*Room
	Register       chan *Client
	unregister     chan *Client
	broadcast      chan Message
	registerLock   *sync.RWMutex
	unregisterLock *sync.RWMutex
	broadcastLock  *sync.RWMutex
}

type Message struct {
	RoomNumber string `json:"room_number"`
	UserName   string `json:"username"`
	Cmd        string `json:"cmd"`
	Payload    string `json:"payload"`
}

type Client struct {
	//cmd    string
	RoomNumber string
	UserName   string
	Password   string
	Send       chan Message
	Conn       *websocket.Conn
	Stop       chan bool
}

type Room struct {
	Clients  map[*Client]*Client
	RoomName string
	Password string
}

var Manager = ConversationManager{
	broadcast:      make(chan Message),
	Register:       make(chan *Client),
	unregister:     make(chan *Client),
	Rooms:          make(map[string]*Room),
	registerLock:   new(sync.RWMutex),
	unregisterLock: new(sync.RWMutex),
	broadcastLock:  new(sync.RWMutex),
}

func (manager *ConversationManager) Start() {
	for {
		select {
		case client := <-manager.Register:
			// 新客户端链接
			manager.registerLock.Lock()
			if _, ok := manager.Rooms[client.RoomNumber]; !ok {
				manager.Rooms[client.RoomNumber] = &Room{
					Clients:  make(map[*Client]*Client),
					Password: client.Password,
				}
			}
			// 塞入房间初次数据
			manager.Rooms[client.RoomNumber].Clients[client] = client
			go func() {
				names := ""
				for key, _ := range manager.Rooms[client.RoomNumber].Clients {
					names += "[ " + key.UserName + " ], "
				}
				names = "<span class='is-inline-block'>" + strings.TrimSuffix(names, ", ") + "</span>"
				manager.broadcast <- Message{
					UserName:   client.UserName,
					Payload:    constant.JoinSuccess + constant.Online + names,
					RoomNumber: client.RoomNumber,
					Cmd:        constant.CmdJoin,
				}
			}()
			manager.registerLock.Unlock()

		case client := <-manager.unregister:
			// 客户端断开链接
			manager.unregisterLock.Lock()
			err := client.Conn.Close()
			if err != nil {
				return
			}
			if _, ok := manager.Rooms[client.RoomNumber]; ok {
				delete(manager.Rooms[client.RoomNumber].Clients, client)
				if len(manager.Rooms[client.RoomNumber].Clients) == 0 {
					delete(manager.Rooms, client.RoomNumber)
				}
				//client.stop <- true
				safeClose(client.Send)

				if manager.Rooms != nil && len(manager.Rooms) != 0 && manager.Rooms[client.RoomNumber] != nil && client.RoomNumber != "" {
					for c, _ := range manager.Rooms[client.RoomNumber].Clients {
						names := ""
						for key, _ := range manager.Rooms[client.RoomNumber].Clients {
							names += "[ " + key.UserName + " ], "
						}
						names = strings.TrimSuffix(names, ", ")
						names = "<span class='is-inline-block'>" + strings.TrimSuffix(names, ", ") + "</span>"
						c.Send <- Message{
							UserName:   client.UserName,
							Payload:    constant.ExitSuccess + constant.Online + names,
							RoomNumber: client.RoomNumber,
							Cmd:        constant.CmdExit,
						}
					}
				}
			}
			manager.unregisterLock.Unlock()

		case message := <-manager.broadcast:
			// 广播消息
			manager.broadcastLock.RLock()
			if manager.Rooms != nil && len(manager.Rooms) != 0 && manager.Rooms[message.RoomNumber] != nil && message.RoomNumber != "" {
				for c, _ := range manager.Rooms[message.RoomNumber].Clients {
					if c != nil && c.Conn != nil && c.Send != nil {
						c.Send <- message
					}
				}
			}
			manager.broadcastLock.RUnlock()
		}

	}
}

func safeClose(ch chan Message) {
	defer func() {
		if recover() != nil {
			log.Println("ch is closed")
		}
	}()
	close(ch)
	log.Println("ch closed successfully")
}
