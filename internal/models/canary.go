package models

import (
	"gorm.io/gorm"
)

type Canary struct {
	gorm.Model
	Id       string
	Name     string
	Type     string
	Redirect string
	User     *User `gorm:"embedded"`
}
