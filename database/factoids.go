package database

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgresql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Remark struct {
	gorm.Model
	Protected    bool
	TriggerCount int
	Text         string `gorm:"index"`
	Retorts      []Retort
}

type Retort struct {
	gorm.Model
	TriggerCount int
	Text         string `gorm:"index"`
}
