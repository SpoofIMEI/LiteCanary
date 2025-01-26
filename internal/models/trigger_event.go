package models

import (
	"time"

	"gorm.io/gorm"
)

type TriggerEvent struct {
	gorm.Model
	Timestamp        time.Time
	Canaryid         string
	Useragent        string
	Keyboardlanguage string
	Ip               string
}
