package conversation

import (
	"encoding/json"
	"log"
	"minichat/constant"
	"minichat/util"
)

func (c *Client) Read() {
	defer func() {
		Manager.unregister <- c
	}()

	for {
		message, err := util.SocketReceive(c.Conn)
		if err != nil {
			return
		}
		Manager.broadcast <- Message{
			UserName:   c.UserName,
			RoomNumber: c.RoomNumber,
			Payload:    string(message),
			Cmd:        constant.CmdChat,
		}
	}
}

func (c *Client) Write() {
	for {
		select {
		case message, isOpen := <-c.Send:
			if !isOpen {
				log.Printf("chan is closed")
				return
			}

			byteData, err := json.Marshal(message)
			if err != nil {
				log.Printf("json marshal error, error is %+v", err)
			} else {
				err = util.SocketSend(c.Conn, byteData)
				if err != nil {
					log.Printf("ocket send error, error is %+v", err)
					return
				}
			}
		case makeStop := <-c.Stop:
			if makeStop {
				break
			}
		}
	}
}
