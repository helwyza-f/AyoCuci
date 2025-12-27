package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"BackendFramework/internal/config"
)

type InfobipService struct {
	APIKey  string
	BaseURL string
	Sender  string
}

func NewInfobipService() *InfobipService {
	return &InfobipService{
		APIKey:  config.INFOBIP_API_KEY,
		BaseURL: "https://api.infobip.com",
		Sender:  config.INFOBIP_SENDER,
	}
}

type InfobipWhatsAppRequest struct {
	Messages []InfobipMessage `json:"messages"`
}

type InfobipMessage struct {
	From    string                `json:"from"`
	To      string                `json:"to"`
	Content InfobipMessageContent `json:"content"`
}

type InfobipMessageContent struct {
	Text string `json:"text"`
}

type InfobipResponse struct {
	Messages []struct {
		MessageID string `json:"messageId"`
		Status    struct {
			GroupID     int    `json:"groupId"`
			GroupName   string `json:"groupName"`
			ID          int    `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"status"`
		To string `json:"to"`
	} `json:"messages"`
}

func (s *InfobipService) SendWhatsAppOTP(phoneNumber, otpCode string) error {
	// Validasi sender dan API key
	if s.Sender == "" {
		return fmt.Errorf("INFOBIP_SENDER tidak diset. Cek config/infobip.go dan .env")
	}
	if s.APIKey == "" {
		return fmt.Errorf("INFOBIP_API_KEY tidak diset. Cek config/infobip.go dan .env")
	}

	// Konversi nomor HP ke format internasional
	if len(phoneNumber) > 0 && phoneNumber[0] == '0' {
		phoneNumber = "62" + phoneNumber[1:]
	}

	message := fmt.Sprintf(
		"🔐 *Kode Verifikasi Anda*\n\n"+
			"Kode OTP: *%s*\n\n"+
			"Kode ini berlaku selama 5 menit.\n"+
			"Jangan bagikan kode ini kepada siapa pun.\n\n"+
			"Jika Anda tidak meminta kode ini, abaikan pesan ini.",
		otpCode,
	)

	payload := InfobipWhatsAppRequest{
		Messages: []InfobipMessage{
			{
				From: s.Sender,
				To:   phoneNumber,
				Content: InfobipMessageContent{
					Text: message,
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest(
		"POST",
		s.BaseURL+"/whatsapp/1/message/text",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", "App "+s.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("infobip API error (status %d): %s", resp.StatusCode, string(body))
	}

	var infobipResp InfobipResponse
	if err := json.Unmarshal(body, &infobipResp); err != nil {
		return fmt.Errorf("error parsing response: %v", err)
	}

	if len(infobipResp.Messages) == 0 {
		return fmt.Errorf("no messages in response")
	}

	return nil
}