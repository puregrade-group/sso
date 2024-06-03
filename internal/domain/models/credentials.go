package models

import (
	"net"
	"time"
)

type Credentials struct {
	Email    string
	Password string
}

type UserCredentials struct {
	Id           uint64 `db:"id"`
	Email        string `db:"email"`
	PasswordHash []byte `db:"pass_hash"`
}

type RefreshToken struct {
	Value     string
	UserId    uint64
	ExpiresIn time.Time
	CreatedBy net.IP
	CreatedAt time.Time
	RevokedBy net.IP
	RevokedAt time.Time
}
