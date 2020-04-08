package main

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

const (
	SslModeDisable = "disable"
	SslModeEnable  = "enable"
	GET            = "get"
	LIST           = "list"
	LIST_BY_CAT    = "listByCat"
	DELETE         = "delete"
	INSERT         = "insert"
	UPDATE         = "update"
)

var stmtMap = map[string]string{
	"get":       "select * from \"user\" where id=$1",
	"listByCat": "select * from \"user\" where category=$1",
	"list":      "select * from \"user\"",
	"insert":    "insert into \"user\" ( name, latitude, longitude, h3index, category) values( $1, $2, $3, $4, $5) returning id",
	"update":    "update \"user\" set latitude=$2, longitude=$3, h3index=$4 where id=$1 returning id, name, latitude, longitude, h3index, category",
	"delete":    "delete from \"user\" where id=$1 returning id, name, latitude, longitude, h3index, category",
}

type SqlDb struct {
	db *sql.DB
}

// NewDb construct a new Db type.
func NewDb(user, password, dbName, sslMode string) (Storage, error) {
	db, err := sql.Open("postgres",
		fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s",
			user, password, dbName, sslMode))
	if err != nil {
		return nil, err
	}

	return &SqlDb{db: db}, nil
}

// GetUser get an user by its id.
func (db *SqlDb) GetUser(id int) (*User, error) {
	stmt, err := db.db.Prepare(stmtMap[GET])
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var user User
	err = stmt.QueryRow(id).Scan(&user.Id, &user.Name, &user.GeoCord.Latitude,
		&user.GeoCord.Longitude, pq.Array(&user.H3Positions), &user.Category)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// AddUser Add an user to the data base.
func (db *SqlDb) AddUser(user *User) (int, error) {
	stmt, err := db.db.Prepare(stmtMap[INSERT])
	if err != nil {
		return 0, err
	}

	defer stmt.Close()

	var id int
	err = stmt.QueryRow(user.Name, user.GeoCord.Latitude, user.GeoCord.Longitude, pq.Array(user.H3Positions), user.Category).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// ListUsers get all users by a category. Returns all users if category="GENERIC".
func (db *SqlDb) ListUsers(category string) (userList []*User, err error) {
	var stmt *sql.Stmt
	var rowIter *sql.Rows

	switch category {
	case Client, ServiceProvider:
		stmt, err = db.db.Prepare(stmtMap[LIST_BY_CAT])
		if err != nil {
			return nil, err
		}

		rowIter, err = stmt.Query(category)

	case Generic:
		stmt, err = db.db.Prepare(stmtMap[LIST])
		if err != nil {
			return nil, err
		}

		rowIter, err = stmt.Query()
	}
	defer stmt.Close()
	if err != nil {
		return nil, err
	}

	for rowIter.Next() {
		var user User
		err = rowIter.Scan(&user.Id, &user.Name, &user.GeoCord.Latitude,
			&user.GeoCord.Longitude, pq.Array(&user.H3Positions), &user.Category)
		if err != nil {
			return nil, err
		}
		userList = append(userList, &user)
	}
	return
}

func (db *SqlDb) DeleteUser(id int) (*User, error) {
	stmt, err := db.db.Prepare(stmtMap[DELETE])
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var user User
	err = stmt.QueryRow(id).Scan(&user.Id, &user.Name, &user.GeoCord.Latitude,
		&user.GeoCord.Longitude, pq.Array(&user.H3Positions), &user.Category)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser update the geo-position of an user.
func (db *SqlDb) UpdateUser(id int, latitude, longitude float64, h3Positions []int64) (*User, error) {
	stmt, err := db.db.Prepare(stmtMap[UPDATE])
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var user User

	err = stmt.QueryRow(id, latitude, longitude, pq.Array(h3Positions)).Scan(&user.Id, &user.Name, &user.GeoCord.Latitude,
		&user.GeoCord.Longitude, pq.Array(&user.H3Positions), &user.Category)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Close close the database
func (db *SqlDb) Close() error {
	return db.db.Close()
}
