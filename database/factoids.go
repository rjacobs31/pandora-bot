package database

import (
	"errors"
	"math/rand"

	"github.com/jinzhu/gorm"
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

type FactoidManager struct {
	db *gorm.DB
}

func initialiseFactoidManager(db *gorm.DB) FactoidManager {
	return FactoidManager{db: db}
}

func (fm *FactoidManager) Add(remark, retort string) (err error) {
	var rem Remark
	fm.db.FirstOrInit(&rem, Remark{Text: remark})

	for _, ret := range rem.Retorts {
		if ret.Text == retort {
			return errors.New("Retort already exists")
		}
	}

	rem.Retorts = append(rem.Retorts, Retort{Text: retort})
	fm.db.Save(&rem)
	return
}

func (fm *FactoidManager) Select(remark string) (retort Retort, err error) {
	var rem Remark
	fm.db.Where("text = ?", remark).First(&rem)

	if len(rem.Retorts) == 0 {
		return
	}

	retort = rem.Retorts[rand.Intn(len(rem.Retorts))]
	return
}
