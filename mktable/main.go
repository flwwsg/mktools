package main

import (
	"fmt"

	"gitee.com/flwwsg/utils-go/errors"
)

func main() {
	var config = Config{
		Host:     "192.168.112.130",
		Port:     3306,
		Socket:   "",
		Username: "dev",
		Password: "123456",
		Debug:    true,
		DBName:   "dump-db",
	}
	repo := NewRepo(&config)
	// get db
	db, err := repo.GetDB(&DB{Name: config.DBName})
	errors.PanicOnErr(err)
	fmt.Printf("\n%v", db)
	fmt.Printf("\n%v\n", db.Tables)
}
