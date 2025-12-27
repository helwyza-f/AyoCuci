package middleware

import (
	"bytes"
	"io"
	"context"
	"time"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"BackendFramework/internal/database"
)

func LogUserActivity() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDInterface, exists := c.Get("userID")
		
		var userID interface{}
		if exists {
			userID = userIDInterface
		} else {
			userID = c.PostForm("user_id")
			if userID == "" {
				userID = "anonymous"
			}
		}

		queryParams := c.Request.URL.Query()
		var requestBody map[string]interface{}
		
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				
				// Try parse as JSON
				if err := bson.UnmarshalExtJSON(bodyBytes, true, &requestBody); err != nil {
					// Jika bukan JSON, coba parse form data
					c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
					if err := c.Request.ParseForm(); err == nil {
						formData := make(map[string]interface{})
						for key, values := range c.Request.PostForm {
							if len(values) == 1 {
								formData[key] = values[0]
							} else {
								formData[key] = values
							}
						}
						requestBody = formData
					}
				}
			}
		}

		go func() {
			logCollection := database.DbAuth.Collection("user_activity")
			
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			
			_, err := logCollection.InsertOne(ctx, bson.M{
				"user_id":      userID,
				"endpoint":     c.Request.URL.Path,
				"method":       c.Request.Method,
				"ip_address":   c.ClientIP(),
				"user_agent":   c.GetHeader("User-Agent"),
				"query_params": queryParams,
				"request_body": requestBody,
				"timestamp":    time.Now(),
			})
			if err != nil {
				fmt.Printf("⚠️ Failed to log user activity: %v\n", err)
			}
		}()
		c.Next()
	}
}