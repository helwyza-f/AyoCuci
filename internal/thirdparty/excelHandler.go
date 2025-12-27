package thirdparty

import(
	"strconv"

	"github.com/xuri/excelize/v2"
	
	"BackendFramework/internal/middleware"

)

type Header struct {
	Text  string  // Header text (e.g., "ID", "Name").
	Width float64 // Column width for the header.
}

func GenerateExcelFile(excelHeader []Header, excelData []map[string]interface{}, sheetName,savePath string) bool {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			middleware.LogError(err,"Failed To Close Excel File")
			// log.Printf("Error closing Excel file: %v", err)
		}
	}()
	// Customize header cell style (e.g., bold font).
	BoldStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	if err != nil {
		middleware.LogError(err,"Failed To Close Excel File")
		// return err
		return false
	}
	// Create a new sheet.
	index, err := f.NewSheet(sheetName)
	if err != nil {
		middleware.LogError(err,"Failed To Create Sheet")
		return false
	}

	// Write excelHeader to the first row and set column widths.
	for colIndex, header := range excelHeader {
		colName, _ := excelize.ColumnNumberToName(colIndex + 1) // e.g., "A", "B", etc.
		cell := colName + "1"                                  // e.g., "A1", "B1", etc.

		// Write header text.
		f.SetCellValue(sheetName, cell, header.Text)

		f.SetCellStyle(sheetName, cell, cell, BoldStyle)

		// Set column width for the header.
		f.SetColWidth(sheetName, colName, colName, header.Width)
	}

	// Write excelData starting from the second row.
	i := 2
	for _, row := range excelData {
		for colIndex, header := range excelHeader {
			colName, _ := excelize.ColumnNumberToName(colIndex + 1)
			cell := colName + strconv.Itoa(i) // e.g., "A2", "B2", etc.
			f.SetCellValue(sheetName, cell,  row[header.Text])
		}
		i++
	}
	// Set the active sheet.
	f.SetActiveSheet(index)


    // Delete the default "Sheet1".
    if sheetName != "Sheet1" {
	    if err := f.DeleteSheet("Sheet1"); err != nil {
	        // return err
			middleware.LogError(err,"Failed To Delete Default Sheet")
			return false
	    }
	}

	// Save the file.
	if err := f.SaveAs(savePath); err != nil {
		middleware.LogError(err,"Failed To Save File")
		return false
	}

	return true
}

// ReadExcelFile reads data from an Excel file and returns headers and data.
func ReadExcelFile(sheetName,filePath string) ([]Header, []map[string]interface{}, bool) {
    f, err := excelize.OpenFile(filePath)
    if err != nil {
		middleware.LogError(err,"Failed To open Excel File")
        return nil, nil, false
    }
    defer func() {
        if err := f.Close(); err != nil {
			middleware.LogError(err,"Failed To Close Excel File")
            // log.Printf("Error closing Excel file: %v", err)
        }
    }()

    // Get all the rows in the first sheet.
    rows, err := f.GetRows(sheetName)
    if err != nil {
		middleware.LogError(err,"Failed to get rows from excel file")
        return nil, nil, false
    }

    // Extract headers and their widths.
    var headers []Header
    for colIndex, headerText := range rows[0] {
        colName, _ := excelize.ColumnNumberToName(colIndex + 1)
        width, _ := f.GetColWidth(sheetName, colName)
        headers = append(headers, Header{
            Text:  headerText,
            Width: width,
        })
    }

    // Extract data starting from the second row.
    var data []map[string]interface{}
    for rowIndex := 1; rowIndex < len(rows); rowIndex++ {
        rowData := make(map[string]interface{})
        for colIndex, header := range headers {
            rowData[header.Text] = rows[rowIndex][colIndex]
        }
        data = append(data, rowData)
    }

    return headers, data, true
}