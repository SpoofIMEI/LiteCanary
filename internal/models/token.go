package models

import (
	"time"

	"gorm.io/gorm"
)

type Token struct {
	gorm.Model
	Id         string
	User       *User `gorm:"embedded"`
	Lastactive time.Time
}
