package main

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type DatabaseConn struct {
	User, Password string
	Host, Database string
	TableName string
	conn *sql.DB
}

func (db *DatabaseConn) InitConnection(User, Pass, Host, Database, table string) {
	conn, err := GetDatabaseConn(GenerateConnStr(db.User, db.Password, db.Host, db.Database))
	if err != nil {
		// todo: default to an in-memory buffer
		panic(err)
	}
	db.User = User
	db.Password = Pass
	db.Host = Host
	db.Database = Database
	db.TableName = table
	db.conn = conn
}

func (db *DatabaseConn) InsertMapping(ip, mac string) {
	query := fmt.Sprintf("INSERT INTO %s (ip, mac) VALUES ($1, $2)", db.TableName)
	_, err := db.conn.Exec(query, ip, mac)
	if err != nil {
		pqError, _ := err.(*pq.Error);
		switch pqError.Code.Name() {
		case "unique_violation":
			db.conn.Exec("UPDATE ipmapping SET ip = $1 WHERE mac = $2", ip, mac)
		default:
			fmt.Printf("[ERROR] %s\n", err)
		}
	} else {
		fmt.Printf("[INSERTED] %s <%s>\n", ip, mac)
	}
}