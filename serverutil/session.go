package serverutil

import "time"

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

type Storage interface {
	Get(string) (Session, error)
	Set(Session, time.Time) error
	Create(interface{}) (Sessoin error)
	Del(string) error
	ClearExpired()
}

type StorageError struct {
	error
	err_msg string
}

func (self *StorageError) Error() string {
	return self.err_msg
}

type StorageNotFoundError struct {
	StorageError
}

type MapSession struct {
	Session
	session map[string]interface{}
	updated bool
	id      string
	expired time.Time
}

func (self *MapSession) Clear() {
	self.session = make(map[string]interface{})
	self.updated = false
}

func (self *MapSession) SetId(key string) {
	self.id = key
}

func (self MapSession) GetId() string {
	return self.id
}

func (self MapSession) Get(key string) interface{} {
	return self.session[key]
}

func (self *MapSession) Set(key string, value interface{}) error {
	self.updated = true
	self.session[key] = value
	return nil
}

func (self *MapSession) All() map[string]interface{} {
	return self.session
}

func (self *MapSession) Del(key string) {
	_, ok := self.session[key]
	self.updated = self.updated || ok
	delete(self.session, key)
}

func (self *MapSession) IsUpdated() bool {
	return self.updated
}

func (self *MapSession) IsEmpty() bool {
	return len(self.session) == 0 && self.updated == false
}

func NewMapSession() *MapSession {
	return &MapSession{
		session: make(map[string]interface{}),
	}
}
