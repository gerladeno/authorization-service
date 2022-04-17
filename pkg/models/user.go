package models

import "time"

type User struct {
	UUID     string
	Username string
	Phone    string
	Created  time.Time
	Updated  time.Time
}
