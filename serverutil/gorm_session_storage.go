package serverutil

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type GormStorage struct {
	Storage
	db *gorm.DB
}

type SessionData map[string]interface{}

func (data *SessionData) Scan(value interface{}) error {
	var err error
	switch value := value.(type) {
	case []byte:
		err = json.Unmarshal(value, data)
	case string:
		err = json.Unmarshal([]byte(value), data)
	default:
		err = errors.New("type not support")
	}
	return err
}

func (data SessionData) Value() (driver.Value, error) {
	return json.Marshal(data)
}

type SessionModel struct {
	ID          uint        `gorm:"primaryKey"`
	Data        SessionData `gorm:"type:text"`
	ExpiredTime time.Time   `gorm:"index,sort:asc,not null"`
}

func NewGormStorage(db *gorm.DB) (*GormStorage, error) {
	err := db.AutoMigrate(&SessionModel{})
	if err != nil {
		return nil, err
	}
	storage := &GormStorage{
		db: db,
	}
	storage.ClearExpired()
	return storage, nil
}

func (self GormStorage) Get(session_id string) (Session, error) {
	var session SessionModel
	tx := self.db.First(&session, "id = ?", session_id)
	if session.ExpiredTime.Before(time.Now()) {
		return nil, errors.New("session " + session_id + " expired")
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	if session.Data == nil {
		return nil, &StorageNotFoundError{}
	}
	id := strconv.Itoa(int(session.ID))
	return &MapSession{
		session: session.Data,
		id:      id,
	}, nil
}

func (self *GormStorage) Set(session Session, expired_time time.Time) error {
	model := &SessionModel{
		Data:        session.All(),
		ExpiredTime: expired_time,
	}
	if session.GetId() != "" {
		sid, err := strconv.Atoi(session.GetId())
		if err != nil {
			return err
		}
		model.ID = uint(sid)
	}
	tx := self.db.Create(model)
	if tx.Error != nil {
		db := self.db.Model(model)
		if !session.IsUpdated() {
			db = db.Select("expired_time")
		}
		tx := db.Updates(model)
		if tx.Error != nil {
			log.Println(tx.Error)
			return tx.Error
		}
	}
	session.SetId(strconv.Itoa(int(model.ID)))
	return nil
}

func (self *GormStorage) ClearExpired() {
	self.db.Delete(&SessionModel{}, "expired_time < ?", time.Now())
}

func (self *GormStorage) Del(session_id string) error {
	tx := self.db.Delete(&SessionModel{}, "id = ?", session_id)
	return tx.Error
}
