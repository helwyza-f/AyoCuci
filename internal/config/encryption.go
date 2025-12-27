package config

import (
	"os"
	"time"
)

var(
	JWT_SIGNATURE_KEY	string	
	ENCRYPTION_KEY	string	
)

const (
	AccessTokenExpiry  = time.Hour * 24
	RefreshTokenExpiry = time.Hour * 24 * 7
)

func InitEncryptionVars() {
	JWT_SIGNATURE_KEY = os.Getenv("JWT_SIGNATURE_KEY"+Prefix)
	ENCRYPTION_KEY = os.Getenv("ENCRYPTION_KEY"+Prefix)
}