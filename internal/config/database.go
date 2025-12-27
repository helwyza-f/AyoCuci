package config

import (
	"os"
)

var(
	// Credential DB Akademik
	DB_CORE_HOSTNAME	string 
	DB_CORE_USERNAME	string 
	DB_CORE_PASSWORD	string 
	DB_CORE_DBNAME	string 


	// Credential DB AUTH
	DB_AUTH_HOSTNAME	string
	DB_AUTH_USERNAME	string
	DB_AUTH_PASSWORD	string
	DB_AUTH_DBNAME		string
)

func InitDatabaseVars() {
	DB_CORE_HOSTNAME	= os.Getenv("DB_CORE_HOSTNAME"+Prefix)
	DB_CORE_USERNAME	= os.Getenv("DB_CORE_USERNAME"+Prefix)
	DB_CORE_PASSWORD	= os.Getenv("DB_CORE_PASSWORD"+Prefix)
	DB_CORE_DBNAME		= os.Getenv("DB_CORE_DBNAME"+Prefix) 

	DB_AUTH_HOSTNAME	=os.Getenv("DB_AUTH_HOSTNAME"+Prefix) 
	DB_AUTH_USERNAME	=os.Getenv("DB_AUTH_USERNAME"+Prefix) 
	DB_AUTH_PASSWORD	=os.Getenv("DB_AUTH_PASSWORD"+Prefix) 
	DB_AUTH_DBNAME		=os.Getenv("DB_AUTH_DBNAME"+Prefix)
}