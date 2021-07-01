package main

import (
	"github.com/a180825/sharedrop-go-server/src"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/auth", src.PeerAuth)
	r.GET("/rooms/:roomId/users/:uid", src.JsonApi)
	r.Run(":10010")
}
