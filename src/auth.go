package src

import (
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
)

func PeerAuth(c *gin.Context) {
	ip := c.ClientIP()
	id := uuid.NewV4().String()

	c.JSON(200, gin.H{
		"uid":       id,
		"public_ip": ip,
	})
}
