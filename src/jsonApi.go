package src

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func JsonApi(c *gin.Context) {
	//升级get请求为webSocket协议
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("error get connection: %v", err)
		log.Fatal(err)
	}
	defer ws.Close()

	var autoInfo AuthInfo
	err = ws.ReadJSON(&autoInfo)
	if err != nil {
		log.Printf("error read json: %v", err)
		c.Error(err)
		return
	}

	roomId := c.Param("roomId")
	uid := c.Param("uid")

	err = roomService.AddUser(roomId, uid, ws, &autoInfo)
	if err != nil {
		log.Printf("error Add user to room: %v", err)
		c.Error(err)
		return
	}
	defer roomService.RemoveUser(uid)
	var data struct {
		Uid     string                 `json:"uid"`
		Message map[string]interface{} `json:"message"`
	}

	for {
		err = ws.ReadJSON(&data)
		if err != nil {
			log.Printf("error read json: %v", err)
			return
		}
		roomService.SendMessage(uid, data.Uid, data.Message)
	}
}
