package serverutil

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type SQLStorageConfig struct {
	TableName string
}

type SQLStorage struct {
	Storage
	db         *sql.DB
	get_sql    *sql.Stmt
	set_sql    *sql.Stmt
	insert_sql *sql.Stmt
	del_sql    *sql.Stmt
}

func NewSQLStorage(db *sql.DB, config SQLStorageConfig) (*SQLStorage, error) {
	db.Exec(fmt.Sprintf("Create table `%s` if not exists (`id` INT NOT NULL, `data` JSON NULL, PRIMARY KEY (`id`)); ", config.TableName))
	get_sql, err := db.Prepare(fmt.Sprintf("select id, data from `%s` where id=? limit 1", config.TableName))
	if err != nil {
		return nil, err
	}
	set_sql, err := db.Prepare(fmt.Sprintf("insert into `%s` (`id`, `data`) values (?, ?) ON DUPLICATE KEY UPDATE data=?", config.TableName))
	if err != nil {
		return nil, err
	}
	insert_sql, err := db.Prepare(fmt.Sprintf("insert into `%s` (`data`) values (?)", config.TableName))
	if err != nil {
		return nil, err
	}
	del_sql, err := db.Prepare(fmt.Sprintf("delete from `%s` where id = ?", config.TableName))
	if err != nil {
		return nil, err
	}
	storage := &SQLStorage{
		db:         db,
		get_sql:    get_sql,
		set_sql:    set_sql,
		insert_sql: insert_sql,
		del_sql:    del_sql,
	}
	return storage, nil
}

func (self *SQLStorage) Get(session_id string) (Session, error) {
	result, err := self.get_sql.Query(session_id)
	if err != nil {
		return nil, err
	}
	if !result.Next() {
		return nil, &StorageNotFoundError{}
	}
	var sid string
	var data []byte
	err = result.Scan(&sid, &data)
	if err != nil {
		return nil, err
	}
	body := make(map[string]interface{})
	err = json.Unmarshal(data, &body)
	if err != nil {
		return nil, err
	}
	return &MapSession{
		session: body,
		id:      sid,
	}, nil
}

func (self *SQLStorage) Set(session Session) error {
	data, err := json.Marshal(session.All())
	if err != nil {
		return err
	}
	if session.GetId() == "" {
		result, err := self.insert_sql.Exec(data)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		sid := fmt.Sprintf("%d", id)
		session.SetId(sid)
	} else {
		_, err = self.set_sql.Exec(
			session.GetId(),
			data,
			data,
		)
		return err
	}
	return nil
}

func (self *SQLStorage) ClearExpired() {
}

func (self *SQLStorage) Del(session_id string) error {
	_, err := self.del_sql.Exec(session_id)
	return err
}
