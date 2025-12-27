package thirdparty

import(
	"github.com/skip2/go-qrcode"
	
	"BackendFramework/internal/middleware"
)

func GenerateQrFile(qrContent, outputPath string) bool {
	qr, err := qrcode.New(qrContent,qrcode.Medium)

	if err != nil {
		middleware.LogError(err,"Failed To Create Qr Generator")
		return false
	}

	err = qr.WriteFile(128,outputPath)

	if err != nil {
		middleware.LogError(err,"Failed To Create Qr Code")
		return false
	}

	return true
}