package config

import "os"

var (
	LDAP_SERVER   = os.Getenv("LDAP_SERVER")    
	LDAP_PORT     = 389
	LDAP_BASE_DN  = os.Getenv("LDAP_BASE_DN")  
)