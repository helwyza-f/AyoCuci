package model

import (
	"time"
)

type TokenData struct {
	UserId                string    `bson:"user_id"`
	LastIpAddress         string    `bson:"last_ip_address"`
	LastUserAgent         string    `bson:"last_user_agent"`
	AccessToken           string    `bson:"access_token"`
	RefreshToken          string    `bson:"refresh_token"`
	RefreshTokenExpiredAt time.Time `bson:"refresh_token_expired"`
	LastLogin             time.Time `bson:"last_login"`
	IsValidToken          string    `bson:"is_valid_token"`
	IsRememberMe          string    `bson:"is_remember_me"`
}

type UserActivity struct {
	Userid    string    `bson:"user_id"`
	Endpoint  string    `bson:"endpoint"`
	Method    string    `bson:"method"`
	Timestamp time.Time `bson:"timestamp"`
}
