package data

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"locationMicroService/libs/actors"
	"locationMicroService/libs/core"
	"log"
)

const (
	GET                = "get"
	GET_CLOSE_WITH_CAT = "getCloseWithCat"
	GET_CLOSE          = "getClose"
	LIST               = "list"
	LIST_BY_CAT        = "listByCat"
	DELETE             = "delete"
	INSERT             = "insert"
	UPDATE             = "update"
)

const (
	DB_USER    = "DB_USER"
	DB_PASS    = "DB_PASS"
	DB_NAME    = "DB_NAME"
	DB_HOST    = "DB_HOST"
	DB_SSLMODE = "DB_SSL_MODE"
)

type stmtConfig struct {
	stmt  *sql.Stmt
	query string
}

type Data struct {
	db      *sql.DB
	stmtMap map[string]*stmtConfig
}

// NewDb construct a new Db type.
func NewDb(user, password, dbHost, dbName, sslMode string) (core.Storage, error) {
	log.Print("Creating Database")
	db, err := sql.Open("postgres",
		fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=%s",
			user, password, dbHost, dbName, sslMode))
	if err != nil {
		return nil, err
	}
	data := &Data{
		db: db,
		stmtMap: map[string]*stmtConfig{
			GET:                {query: "select * from \"user\" where id=$1"},
			GET_CLOSE_WITH_CAT: {query: "select * from \"user\" where h3index[$1]=$2 and category=$3"},
			GET_CLOSE:          {query: "select * from \"user\" where h3index[$1]=$2"},
			LIST_BY_CAT:        {query: "select * from \"user\" where category=$1"},
			LIST:               {query: "select * from \"user\""},
			INSERT:             {query: "insert into \"user\" ( name, latitude, longitude, h3index, category) values( $1, $2, $3, $4, $5) returning id"},
			UPDATE:             {query: "update \"user\" set latitude=$2, longitude=$3, h3index=$4 where id=$1 returning id, name, latitude, longitude, h3index, category"},
			DELETE:             {query: "delete from \"user\" where id=$1 returning id, name, latitude, longitude, h3index, category"},
		},
	}
	for key, stmtConf := range data.stmtMap {
		data.stmtMap[key].stmt, err = data.db.Prepare(stmtConf.query)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

// GetUser get an user by its id.
func (db *Data) GetUser(id int) (*actors.User, error) {
	log.Print("Getting User")

	var user actors.User
	err := db.stmtMap[GET].stmt.QueryRow(id).Scan(&user.Id, &user.Name, &user.GeoCord.Latitude,
		&user.GeoCord.Longitude, pq.Array(&user.H3Positions), &user.Category)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetCloseUsers returns a list of users with the same h3IndexPos given a resolution.
// Can be specified a category or use category = "GENERIC" for all users.
func (db *Data) GetCloseUsers(resolution int, h3IndexPos int64, category string) (userList []*actors.User, err error) {
	log.Print("Getting Close User")

	var rowIter *sql.Rows

	switch category {
	case actors.Client, actors.ServiceProvider:
		rowIter, err = db.stmtMap[GET_CLOSE_WITH_CAT].stmt.Query(resolution+1, h3IndexPos, category)

	case actors.Generic:
		rowIter, err = db.stmtMap[GET_CLOSE].stmt.Query(resolution+1, h3IndexPos)
	}
	if err != nil {
		return nil, err
	}

	for rowIter.Next() {
		var user actors.User
		err = rowIter.Scan(&user.Id, &user.Name, &user.GeoCord.Latitude,
			&user.GeoCord.Longitude, pq.Array(&user.H3Positions), &user.Category)
		if err != nil {
			return nil, err
		}
		userList = append(userList, &user)
	}
	return
}

// AddUser Add an user to the data base.
func (db *Data) AddUser(user *actors.User) (int, error) {
	log.Print("Adding User")

	var id int
	err := db.stmtMap[INSERT].stmt.QueryRow(user.Name, user.GeoCord.Latitude, user.GeoCord.Longitude, pq.Array(user.H3Positions), user.Category).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// ListUsers get all users by a category. Returns all users if category="GENERIC".
func (db *Data) ListUsers(category string) (userList []*actors.User, err error) {
	log.Print("Getting all User")
	var rowIter *sql.Rows

	switch category {
	case actors.Client, actors.ServiceProvider:
		rowIter, err = db.stmtMap[LIST_BY_CAT].stmt.Query(category)

	case actors.Generic:
		rowIter, err = db.stmtMap[LIST].stmt.Query()
	}
	if err != nil {
		return nil, err
	}

	for rowIter.Next() {
		var user actors.User
		err = rowIter.Scan(&user.Id, &user.Name, &user.GeoCord.Latitude,
			&user.GeoCord.Longitude, pq.Array(&user.H3Positions), &user.Category)
		if err != nil {
			return nil, err
		}
		userList = append(userList, &user)
	}
	return
}

// DeleteUser remove an user by its id.
func (db *Data) DeleteUser(id int) (*actors.User, error) {
	log.Print("Removing User")

	var user actors.User
	err := db.stmtMap[DELETE].stmt.QueryRow(id).Scan(&user.Id, &user.Name, &user.GeoCord.Latitude,
		&user.GeoCord.Longitude, pq.Array(&user.H3Positions), &user.Category)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser update the geo-position of an user.
func (db *Data) UpdateUser(id int, latitude, longitude float64, h3Positions []int64) (*actors.User, error) {
	log.Print("Updating User")
	var user actors.User
	err := db.stmtMap[UPDATE].stmt.QueryRow(id, latitude, longitude, pq.Array(h3Positions)).Scan(&user.Id, &user.Name, &user.GeoCord.Latitude,
		&user.GeoCord.Longitude, pq.Array(&user.H3Positions), &user.Category)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Close close the database.
func (db *Data) Close() error {
	log.Print("Closing database")
	// Closing all stataments
	for s := range db.stmtMap {
		if err := db.stmtMap[s].stmt.Close(); err != nil {
			return err
		}
	}
	return db.db.Close()
}
