package thirdparty

import(
	"strings"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	
	"BackendFramework/internal/middleware"
)

func GeneratePdf(pdfBody, outputPath string) bool {
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		middleware.LogError(err,"Failed To Create Pdf Generator")
		return false
	}

	myReader := strings.NewReader(pdfBody)
	pdfg.AddPage(wkhtmltopdf.NewPageReader(myReader))

	pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)

	pdfg.Dpi.Set(300)
	err = pdfg.Create()

	if err != nil {
		middleware.LogError(err,"Pdf Creation Falied")
		return false
	}
	err = pdfg.WriteFile(outputPath)
	
	if err != nil {
		middleware.LogError(err,"Failed To Write PDF File")
		return false
	}
	return true
}