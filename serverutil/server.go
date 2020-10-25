package serverutil

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type ServerSettings struct {
	PrivateKeyPath string
	ServerAddress  string
	Storage        Storage
	Db             *gorm.DB
	SessionName    string
}

func NewSessionHttpServer(config ServerSettings) (*HttpServer, error) {
	jwt, err := NewJwtHelperFromPem(config.PrivateKeyPath)
	if err != nil {
		return nil, err
	}
	handler := &SessionServMux{
		jwt:             jwt,
		session_storage: config.Storage,
		session_name:    config.SessionName,
	}
	server := &HttpServer{
		Server: &http.Server{
			Handler: handler,
			Addr:    config.ServerAddress,
		},
		db: config.Db,
	}
	return server, nil
}

type HttpServer struct {
	*http.Server
	db *gorm.DB
}

func (self HttpServer) DB() *gorm.DB {
	return self.db
}

type SessionServMux struct {
	http.ServeMux
	jwt             *JwtHelper
	session_storage Storage
	session_name    string
}

func (self *SessionServMux) getSession(r *http.Request) Session {
	signature, err := r.Cookie(self.session_name)

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

func (self *SessionServMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session := self.getSession(r)
	wrap := &ResponseWriter{
		ResponseWriter: w,
		Session:        session,
		callback: func(w *ResponseWriter) {
			expire_date := time.Now().AddDate(0, 0, 1)
			self.session_storage.Set(session, expire_date)
			session := map[string]interface{}{
				"sid": w.Session.GetId(),
			}
			data, err := self.jwt.Sign(
				session,
				expire_date,
			)
			if err == nil {
				http.SetCookie(w, &http.Cookie{
					Name:     self.session_name,
					Value:    string(data),
					HttpOnly: true,
					Secure:   false,
					Path:     "/",
					Expires:  expire_date,
				})
			} else {
				log.SetFlags(log.LstdFlags | log.Lshortfile)
				log.Println(err)
			}
		},
	}

	self.ServeMux.ServeHTTP(wrap, r)
	if !wrap.done {
		wrap.WriteHeader(404)
		wrap.Write([]byte("404 not found"))
	}
	if wrap.statusCode == 0 {
		wrap.statusCode = 200
	}
	log.Printf("%s %s %d\n", r.Host, r.URL.Path, wrap.statusCode)
}

type ResponseWriter struct {
	http.ResponseWriter
	Session    Session
	statusCode int
	done       bool
	callback   func(*ResponseWriter)
}

func (self *ResponseWriter) Header() http.Header {
	return self.ResponseWriter.Header()
}

func (self *ResponseWriter) Write(b []byte) (int, error) {
	if self.Session.IsUpdated() && !self.done {
		self.callback(self)
	}
	self.done = true
	return self.ResponseWriter.Write(b)
}

func (self *ResponseWriter) WriteHeader(code int) {
	if self.Session.IsUpdated() && !self.done {
		self.callback(self)
	}
	self.done = true
	self.statusCode = code
	self.ResponseWriter.WriteHeader(code)
}

func (self *ResponseWriter) WriteJson(body interface{}) (int, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}
	return self.Write(data)
}
