package database

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Database struct {
	*gorm.DB

	FactoidManager FactoidManager
}

func InitialiseDB(config Config) (db *Database, err error) {
	gdb, err := gorm.Open(config.Type, config.ConnectionString)
	if err != nil {
		return
	}

	gdb = db.AutoMigrate(&Remark{}, &Retort{})

	db = &Database{
		DB:             gdb,
		FactoidManager: initialiseFactoidManager(gdb),
	}
	return
}
