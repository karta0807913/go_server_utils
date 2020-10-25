package serverutil

import (
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func createServer(t *testing.T, dbname string) (*gin.Engine, *gorm.DB) {
	serv := gin.Default()
	os.Remove(dbname)
	db, err := gorm.Open(sqlite.Open(dbname), &gorm.Config{})
	if err != nil {
		t.Fatal("create sqlite error", err)
	}
	// err = model.InitDB(db)
	if err != nil {
		t.Fatal("init sqlite error", err)
	}
	pKey, err := GenerateKey()
	if err != nil {
		t.Fatal("create pKey error", err)
	}
	jwt, err := NewJwtHelper(pKey)
	if err != nil {
		t.Fatal("create jwt error", err)
	}
	storage, err := NewGormStorage(db)
	if err != nil {
		t.Fatal("create storage error", err)
	}
	serv.Use(
		NewGinSessionFactory(jwt, storage).SessionMiddleware("session"),
	)
	return serv, db
}

func TestGormStorage(t *testing.T) {
	server, db := createServer(t, "test.sqlite")
	defer os.Remove("test.sqlite")
	server.GET("/test", func(c *gin.Context) {
		session := c.MustGet("session").(Session)
		data := session.Get("pkey")
		if data == nil {
			c.AbortWithStatus(403)
		} else {
			c.String(200, session.GetId())
		}
	})
	server.GET("/set", func(c *gin.Context) {
		session := c.MustGet("session").(Session)
		session.Set("pkey", "hi")
		c.String(200, "")
	})
	request := httptest.NewRequest("GET", "/set", nil)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	assert.Equal(t, response.Code, 200)

	request = httptest.NewRequest("GET", "/test", nil)
	cookie := response.Result().Cookies()[0]
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	server.ServeHTTP(response, request)
	assert.Equal(t, response.Code, 200)

	request = httptest.NewRequest("GET", "/test", nil)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	server.ServeHTTP(response, request)
	assert.Equal(t, response.Code, 200)

	var session SessionModel
	assert.Equal(t, db.First(&session,
		"id=?",
		response.Body.String(),
	).Error, nil)
	session.ExpiredTime = time.Now()
	db.Updates(session)

	request = httptest.NewRequest("GET", "/test", nil)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	server.ServeHTTP(response, request)
	assert.Equal(t, response.Code, 403)
}
