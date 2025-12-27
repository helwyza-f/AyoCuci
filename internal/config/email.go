package config

import(
	"os"
	"strconv"
)

var (
	CONFIG_SMTP_HOST 		string
	CONFIG_SMTP_PORT 		int
	CONFIG_SENDER_NAME 		string 
	CONFIG_AUTH_EMAIL 		string
	CONFIG_AUTH_PASSWORD 	string
)

func InitEmailVars() {
	CONFIG_SMTP_HOST			= os.Getenv("CONFIG_SMTP_HOST"+Prefix)
	CONFIG_SMTP_PORT,_			= strconv.Atoi(os.Getenv("CONFIG_SMTP_PORT"+Prefix))
	CONFIG_SENDER_NAME			= os.Getenv("CONFIG_SENDER_NAME"+Prefix)
	CONFIG_AUTH_EMAIL			= os.Getenv("CONFIG_AUTH_EMAIL"+Prefix)
	CONFIG_AUTH_PASSWORD			= os.Getenv("CONFIG_AUTH_PASSWORD"+Prefix)
}