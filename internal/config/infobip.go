package config

import (
	"fmt"
	"os"
)

var (
	INFOBIP_API_KEY string
	INFOBIP_SENDER  string
)

func InitInfobipVars() {
	INFOBIP_API_KEY = os.Getenv("INFOBIP_API_KEY" + Prefix)
	INFOBIP_SENDER = os.Getenv("INFOBIP_SENDER" + Prefix)

	// Debug log - hapus setelah berhasil
	fmt.Println("=== Infobip Config ===")
	fmt.Println("Prefix:", Prefix)
	fmt.Println("Looking for: INFOBIP_SENDER" + Prefix)
	fmt.Println("INFOBIP_SENDER:", INFOBIP_SENDER)
	if len(INFOBIP_API_KEY) > 20 {
		fmt.Println("INFOBIP_API_KEY:", INFOBIP_API_KEY[:20]+"...")
	} else {
		fmt.Println("INFOBIP_API_KEY: [EMPTY or too short]")
	}
	fmt.Println("======================")
}