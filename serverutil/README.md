# A simple gin server with jwt session

Install command

```
go get https://github.com/karta0807913/go_server_utils/serverutil
```

## func NewGinServer

```golang
func NewGinServer(config ServerSettings) (*gin.Engine, error)
```

## type ServerSettings

```golang
type ServerSettings struct {
    // a pem struct private key path
	PrivateKeyPath string
    // server listen address
	ServerAddress  string
	Storage        Storage
	SessionName    string
}
```

## type Storage

```golang
type Storage interface {
	Get(string) (Session, error)
	Set(Session, time.Time) error
	Create(interface{}) (Sessoin error)
	Del(string) error
	ClearExpired()
}
```

## type Session

```golang
type Session interface {
	Get(string) interface{}
	Set(string, interface{}) error
	GetId() string
	SetId(string)
	Del(string)
	All() map[string]interface{}
	IsUpdated() bool
	IsEmpty() bool
	Clear()
}
```

## func NewGormStorage

```golang
func NewGormStorage(db *gorm.DB) (*GormStorage, error)
```

## type GormStorage

```golang
type GormStorage struct {
	Storage
}
```

## type SessionModel

GormStorage will create a table named session_model

```golang
type SessionModel struct {
	ID          uint        `gorm:"primaryKey"`
	Data        SessionData `gorm:"type:text"`
	ExpiredTime time.Time   `gorm:"index,sort:asc,not null"`
}
```

## type SessionData

```golang
type SessionData map[string]interface{}
```

## type GinSessionFactory

```golang
type GinSessionFactory struct {
}
```

## func (*GinSessionFactory) SessionMiddleware

```golang
func (self *GinSessionFactory) SessionMiddleware(sessionName string) gin.HandlerFunc
```

# Example

```golang
package main

import (
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"github.com/gin-gonic/gin"
	"github.com/karta0807913/go_server_utils/serverutil"
)

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("start db error", err)
	}
	storage, err := serverutil.NewGormStorage(db)

	serv := gin.Default()
	_, err = os.Stat("./private.pem")
	if os.IsNotExist(err) {
		pKey, err := serverutil.GenerateKey()
		if err != nil {
			log.Fatal("Generate key error", err)
		}
		serverutil.SavePEMKey(PrivateKeyPath, pKey)
	}
	jwt, err := serverutil.NewJwtHelperFromPem(PrivateKeyPath)
	if err != nil {
		log.Fatal("read private key file error", err)
	}
	storage, err := serverutil.NewGormStorage(db)
	if err != nil {
		log.Fatal("create storage error", err)
	}
	sessionFactory := serverutil.NewGinSessionFactory(jwt, storage)
	serv.Use(sessionFactory.SessionMiddleware("session"))

	serv.Run("0.0.0.0:4000")
}
```