package models

type Permission struct {
	Id       int32
	Resource string
	Action   string
	// Type        string // "base" | "custom"
	Description string
}
