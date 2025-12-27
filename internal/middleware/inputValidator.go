package middleware

import(
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func InitValidator() {
	validate = validator.New()
}

func validateStruct(s interface{}) error {
	return validate.Struct(s)
}

func InputValidator(obj interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBind(obj); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code" : http.StatusBadRequest,
				"error": "Invalid input",
			})
			c.Abort()
			return
		}

		if err := validateStruct(obj); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code" : http.StatusBadRequest,
				"error": err.Error(),
			})
			c.Abort()
			return
		}

		c.Set("validatedInput", obj)
		c.Next()
	}
}

func ValidateFile(maxSize,fileSize int64, fileExt string, allowedTypes []string )(bool, string) {
	var errMsg strings.Builder
	fileStatus := true
	isValidExt := false
	for _, allowedExt := range allowedTypes {
		if fileExt == allowedExt {
			isValidExt = true
		}
	}
	if isValidExt == false {
		fileStatus = false
		extensionsString := strings.Join(allowedTypes, ",")
		outputString := fmt.Sprintf("Uploaded file is not allowed ( %s ) ", extensionsString)
		errMsg.WriteString(outputString)
	}
	if fileSize > maxSize {
		if fileStatus == false {
			errMsg.WriteString("& ")	
		}
		fileStatus = false
		outputString := fmt.Sprintf("Uploaded file size exceeds %d MB", maxSize/(1024*1024))
		errMsg.WriteString(outputString)
	}

	return fileStatus,errMsg.String()
}