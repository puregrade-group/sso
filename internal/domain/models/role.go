package models

type Role struct {
	Id          int32
	Name        string
	Permissions []Permission
	Description string
}
