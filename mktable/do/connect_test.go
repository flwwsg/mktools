package do

import (
	"fmt"
	"testing"

	"gitee.com/flwwsg/utils-go/errors"
)

var config = Config{
	Host:     "192.168.112.130",
	Port:     3306,
	Socket:   "",
	Username: "dev",
	Password: "123456",
	Debug:    true,
}

func TestNewRepo(t *testing.T) {
	var err error
	// var dbs []DB
	// var tables []Table
	// var columns []Column
	var schemas []schema
	dbCond := schema{
		Name:      "team_place",
		CharSet:   "utf8",
		Collation: "utf8_general_ci",
	}
	repo := NewRepo(&config)
	err = repo.engine.Find(&schemas, &dbCond)
	errors.PanicOnErr(err)
	fmt.Printf("xx\n%v\n", schemas)
}
