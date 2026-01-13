// internal/utils/whatsapp.go
package thirdparty

import (
	"net/http"
	"net/url"
	"strings"
)

func SendFonnteNotification(phone string, message string) {
    token := "6FaSp1ad3Fa9tstaBjwi" // Pakai token Anda
    
    apiURL := "https://api.fonnte.com/send"
    payload := url.Values{}
    payload.Add("target", phone)
    payload.Add("message", message)

    req, _ := http.NewRequest("POST", apiURL, strings.NewReader(payload.Encode()))
    req.Header.Add("Authorization", token)
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

    client := &http.Client{}
    client.Do(req)
}