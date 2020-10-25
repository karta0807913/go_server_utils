package serverutil

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type GinSessionFactory struct {
	jwt             *JwtHelper
	session_storage Storage
}

func NewGinSessionFactory(jwt *JwtHelper, storage Storage) *GinSessionFactory {
	return &GinSessionFactory{
		jwt:             jwt,
		session_storage: storage,
	}
}

func (self *GinSessionFactory) final(sessionName string, session Session, c *gin.Context) {
	if session.IsEmpty() {
		return
	}
	expire_date := time.Now().AddDate(0, 0, 1)
	self.session_storage.Set(session, expire_date)
	sid := map[string]interface{}{
		"sid": session.GetId(),
	}
	data, err := self.jwt.Sign(
		sid,
		expire_date,
	)
	if err != nil {
		log.Printf("get error %s when sign a cookie", err.Error())
		return
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     sessionName,
		Value:    string(data),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		Expires:  expire_date,
	})
}

type GinResponseWriter struct {
	gin.ResponseWriter
	Session  Session
	done     bool
	callback func(*GinResponseWriter)
}

func (self *GinResponseWriter) Write(b []byte) (int, error) {
	if self.Session.IsUpdated() && !self.done {
		self.callback(self)
	}
	self.done = true
	return self.ResponseWriter.Write(b)
}

func (self *GinResponseWriter) WriteHeader(code int) {
	if self.Session.IsUpdated() && !self.done {
		self.callback(self)
	}
	self.done = true
	self.ResponseWriter.WriteHeader(code)
}

func (self *GinResponseWriter) WriteHeaderNow() {
	if self.Session.IsUpdated() && !self.done {
		self.callback(self)
	}
	self.done = true
	self.ResponseWriter.WriteHeaderNow()
}

func (self *GinSessionFactory) SessionMiddleware(sessionName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := self.GetSession(c, sessionName)
		wrap := &GinResponseWriter{
			ResponseWriter: c.Writer,
			Session:        session,
			callback: func(w *GinResponseWriter) {
				self.final(sessionName, session, c)
			},
		}
		c.Writer = wrap
		c.Set("session", session)
		c.Next()
	}
}

func (self *GinSessionFactory) GetSession(c *gin.Context, session_name string) Session {
	signature, err := c.Request.Cookie(session_name)
	if err != nil {
		return NewMapSession()
	}
	data := []byte(signature.Value)
	claim, err := self.jwt.Verify(data)
	if err != nil {
		return NewMapSession()
	}
	id, ok := claim.Set["sid"]
	if !ok {
		return NewMapSession()
	}
	sid, ok := id.(string)
	if !ok {
		return NewMapSession()
	}
	session, err := self.session_storage.Get(sid)
	if err != nil {
		log.Printf("fetch session error, reason: %s\n", err.Error())
		return NewMapSession()
	}
	return session
}
