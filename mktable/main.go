package main

import (
	"mktools/mktable/do"
	"os"

	"gitee.com/flwwsg/utils-go/errors"
)

func main() {
	var config = do.Config{
		Host:     "192.168.112.130",
		Port:     3306,
		Socket:   "",
		Username: "dev",
		Password: "123456",
		Debug:    true,
		DBName:   "dump-db",
	}
	repo := do.NewRepo(&config)
	// get db
	db, err := repo.GetDB(&do.DB{Name: config.DBName})
	errors.PanicOnErr(err)
	// fmt.Printf("\n%v", db)
	// fmt.Printf("\n%v\n", db.Tables)
	do.RenderTable(db.Tables, os.Stdout)

}
