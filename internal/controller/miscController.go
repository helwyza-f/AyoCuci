package controller

import(
	"strings"
	"strconv"
	"net/http"
	"time"
	"io/ioutil"
	"path/filepath"
	"os"

	"github.com/gin-gonic/gin"

	"BackendFramework/internal/service"
	"BackendFramework/internal/thirdparty"
	"BackendFramework/internal/middleware"
)

func TryGeneratePdf(c *gin.Context) {
	f, err := ioutil.ReadFile("./web/html/email_template.html")
	if err != nil {
		middleware.LogError(err,"Failed open HTML")
		c.JSON(http.StatusOK, gin.H{
			"code"	: http.StatusInternalServerError,
			"message":     "Failed To Open Html File",
		})
		return 
	}

	year, _, _ := time.Now().Date()
	templateString := string(f)
	templateString = strings.Replace(templateString,"{{nama}}","Test",1)
	templateString = strings.Replace(templateString,"{{Opening_text}}","Test",1)
	templateString = strings.Replace(templateString,"{{keterangan}}","Test",1)
	templateString = strings.Replace(templateString,"{{Year}}",strconv.Itoa(year),1)
	templateString = strings.Replace(templateString,"{{Link}}","http://localhost",1)
	templateString = strings.Replace(templateString,"{{Nama Sistem}}","Back End Framework",1)

	status := thirdparty.GeneratePdf(templateString,"./temp/output.pdf")
	if status == false {
		c.JSON(http.StatusOK, gin.H{
			"code"	: http.StatusInternalServerError,
			"message":     "Failed To Generate Pdf File",
		})
		return 
	}
	c.JSON(http.StatusOK, gin.H{
		"code"	: http.StatusOK,
		"message":     "Check If File is generated properly",
	})
}

func SendMail(c *gin.Context) {
	recipientData := []thirdparty.RecipientStruct{{
		Name:	"Test",
        Email: 	"test@gmail.com",
	}}
	year, _, _ := time.Now().Date()

	var mailSubject = "Test Mail "+strconv.Itoa(year)

	f, err := ioutil.ReadFile("./web/html/email_template.html")
	if err != nil {
		middleware.LogError(err,"Failed open HTML")
		c.JSON(http.StatusOK, gin.H{
			"code"	: http.StatusInternalServerError,
			"message":     "Failed To Open Html File",
		})
		return 
	}

	templateString := string(f)
	templateString = strings.Replace(templateString,"{{nama}}","Test",1)
	templateString = strings.Replace(templateString,"{{Opening_text}}","Test",1)
	templateString = strings.Replace(templateString,"{{keterangan}}","Test",1)
	templateString = strings.Replace(templateString,"{{Year}}",strconv.Itoa(year),1)
	templateString = strings.Replace(templateString,"{{Link}}","http://localhost",1)
	templateString = strings.Replace(templateString,"{{Nama Sistem}}","Back End Framework",1)

	status := thirdparty.SendEmail(templateString,mailSubject,recipientData)

	if status == false {
		c.JSON(http.StatusOK, gin.H{
			"code"	: http.StatusInternalServerError,
			"message":     "Failed To Send Mail",
		})
		return 
	}
	c.JSON(http.StatusOK, gin.H{
		"code"	: http.StatusOK,
		"message":     "Check If Mail Is Sent properly",
	})
}

func GenerateExcel(c *gin.Context) {
	 headers := []thirdparty.Header{
        {Text: "User", Width: 10},
        {Text: "Email", Width: 20},
        {Text: "Usergroup", Width: 30},
    }
    users := service.GetAllUsers()
    var excelData []map[string]interface{}
    for _,user := range users {
    	excelData = append(excelData, map[string]interface{}{
    		"User" : user.NamaLengkap,
    		"Email": user.Email,
    		"Usergroup":user.Group,
    	})
    }
    sheetName := "User List"
    savePath := "temp/users.xlsx"

    status := thirdparty.GenerateExcelFile(headers,excelData,sheetName,savePath)

	if status == false {
		c.JSON(http.StatusOK, gin.H{
			"code"	: http.StatusInternalServerError,
			"message":     "Failed To Generate Excel",
		})
		return 
	}
	c.JSON(http.StatusOK, gin.H{
		"code"	: http.StatusOK,
		"message":     "Check If Excel Is Properly Generated",
	})
} 

func ReadExcel(c *gin.Context) {
	sheetName := c.PostForm("sheetName")
	if sheetName == "" {
		c.JSON(http.StatusOK, gin.H{
			"code" :http.StatusBadRequest,
			"error": "sheet name is requred",
		})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code" :http.StatusBadRequest,
			"error": "file is required",
		})
		return
	}

	const maxFileSize = 3 * 1024 * 1024 // Max file size in bytes (3 MB)
	// Allowed file extensions
	var allowedExtensions = []string{".xlsx"}
	const uploadDir   = "./temp/" // Directory to save uploaded files locally

	// Check file type (by extension)
	ext := strings.ToLower(filepath.Ext(file.Filename))

	fileStatus,errMsg := middleware.ValidateFile(maxFileSize, file.Size,ext,allowedExtensions)
	if(fileStatus == false) {
		c.JSON(http.StatusOK, gin.H{
			"code" :http.StatusBadRequest,
			"error": errMsg,
		})
		return
	}
	// Save the file locally
	localFileName := "temp"+ext
	localFilePath := filepath.Join(uploadDir, localFileName)
	if err := c.SaveUploadedFile(file, localFilePath); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code" :http.StatusInternalServerError,
			"error": "failed to save file locally",
		})
		return
	}
	excelHeader,excelData,status := thirdparty.ReadExcelFile(sheetName,localFilePath)
	if status == false {
		c.JSON(http.StatusOK, gin.H{
			"code"	: http.StatusInternalServerError,
			"message":     "Failed To Read Excel",
		})
		return 
	}
	// Delete the local file after successful upload
	if err := os.Remove(localFilePath); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code" :http.StatusInternalServerError,
			"error": "failed to clean up local file",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code"	: http.StatusOK,
		"message":     "Check If Excel Data Is Right",
		"excelData" : gin.H{
			"header" : excelHeader,
			"data"	: excelData,
		},
	})
}

func PingMongo(c *gin.Context) {
	service.TestPing()
	c.JSON(http.StatusOK, gin.H{
		"code"	: http.StatusOK,
		"message":     "Check Ping",
	})
}