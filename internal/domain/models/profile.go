package models

import "time"

type BriefProfile struct {
	FirstName   string
	LastName    string
	DateOfBirth time.Time
}

type Profile struct {
	BriefProfile
}
