package serverutil

import (
	"github.com/gin-gonic/gin"
)

func NewGinServer(config ServerSettings) (*gin.Engine, error) {
	jwt, err := NewJwtHelperFromPem(config.PrivateKeyPath)
	if err != nil {
		return nil, err
	}
	server := gin.New()
	server.Use(
		NewGinSessionFactory(
			jwt, config.Storage,
		).SessionMiddleware(
			config.SessionName,
		),
	)
	return server, nil
}
