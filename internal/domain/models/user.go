package models

type User struct {
	Id       [16]byte `json:"id" db:"id"` // uuid
	Email    string   `json:"email" db:"email" binding:"required"`
	PassHash []byte   `json:"passHash"`
	Roles    []Role   `json:"roles"`
}
