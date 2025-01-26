package models

import "time"

type LocalTriggerEvent struct {
	Timestamp        time.Time
	Useragent        string
	Keyboardlanguage string
	Ip               string
}

type FilteredCanary struct {
	Name     string
	Id       string
	Type     string
	Redirect string
	History  *[]LocalTriggerEvent
}
