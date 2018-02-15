package database

import (
	"bytes"
	"errors"
	"math/rand"
	"strings"
	"unicode"

	"github.com/jinzhu/gorm"
)

// Remark represents a user message with associated responses.
type Remark struct {
	gorm.Model
	Protected    bool
	TriggerCount int
	Text         string `gorm:"unique_index"`
	Retorts      []Retort
}

// Retort represents a response to registered user messages.
type Retort struct {
	gorm.Model
	RemarkID     uint
	TriggerCount int
	Text         string `gorm:"index"`
}

// FactoidManager handles DB calls for Remarks and Retorts.
type FactoidManager struct {
	db *gorm.DB
}

func initialiseFactoidManager(db *gorm.DB) (fm FactoidManager) {
	fm = FactoidManager{db: db}
	return
}

// cleanTrigger prepares a trigger trimming spaces, converting
// to lowercase, and removing special characters.
func cleanTrigger(trigger string) (out string) {
	simplifiedString := strings.ToLower(strings.TrimSpace(trigger))
	result := bytes.Buffer{}

	// Replace special characters and multiple spaces with a
	// single space.
	prev := ' '
	for _, r := range simplifiedString {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
		} else if unicode.IsSpace(r) || unicode.IsSymbol(r) {
			r = ' '
			if prev != ' ' {
				result.WriteRune(r)
			}
		}

		prev = r
	}

	out = result.String()
	return
}

// Add attempts to register a Retort for a given Remark.
func (fm *FactoidManager) Add(remark, retort string) (err error) {
	rem := Remark{}
	fm.db.FirstOrCreate(&rem, Remark{Text: cleanTrigger(remark)})

	query := Retort{RemarkID: rem.ID, Text: retort}
	ret := Retort{}
	if !fm.db.Where(&query).First(&ret).RecordNotFound() {
		return errors.New("Retort already exists")
	}

	fm.db.Save(&query)
	return
}

// Select attempts to find a random Retort for a given Remark.
func (fm *FactoidManager) Select(remark string) (retort *Retort, err error) {
	simplifiedTrigger := cleanTrigger(remark)
	if simplifiedTrigger == "" {
		return
	}

	query := Remark{Text: simplifiedTrigger}
	rem := Remark{}
	var retorts []Retort
	if fm.db.Where(&query).First(&rem).RecordNotFound() {
		return
	}

	fm.db.Model(&rem).Related(&retorts)
	if len(retorts) == 0 {
		return
	}
	retort = &retorts[rand.Intn(len(retorts))]

	rem.TriggerCount += 1
	fm.db.Save(rem)

	retort.TriggerCount += 1
	fm.db.Save(retort)
	return
}
