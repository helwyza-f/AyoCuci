package middleware

import (
    "context"
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "errors"
    "io"
    "time"
    "net/http"
    "strings"
    "strconv"
    
    "github.com/golang-jwt/jwt/v4"
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/v2/bson"
    
    "BackendFramework/internal/config"
    "BackendFramework/internal/database"
    "BackendFramework/internal/model"
)

var jwtSecret = []byte(config.JWT_SIGNATURE_KEY)

type AccessClaims struct {
    UserID   string `json:"user_id"`
    OutletID uint   `json:"outlet_id"`
    jwt.RegisteredClaims
}

func GenerateAccessToken(userID string, outletID uint) (string, error) {
    claims := &AccessClaims{
        UserID:   userID,
        OutletID: outletID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.AccessTokenExpiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "BackendFramework UIB",
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

func GenerateRefreshToken() (string, error) {
    bytes := make([]byte, 32)
    _, err := rand.Read(bytes)
    if err != nil {
        return "", err
    }
    plainText := base64.StdEncoding.EncodeToString(bytes)
    key := []byte(config.ENCRYPTION_KEY)
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }
    nonce := make([]byte, 12)
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }
    aesGCM, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    cipherText := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)
    return base64.StdEncoding.EncodeToString(cipherText), nil
}

func ValidateToken(tokenString string) (*AccessClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    })
    if err != nil {
        return nil, err
    }
    claims, ok := token.Claims.(*AccessClaims)
    if !ok || !token.Valid {
        return nil, errors.New("invalid token")
    }
    var storedToken model.TokenData
    err = database.DbAuth.Collection("access_tokens").FindOne(context.TODO(), bson.M{"user_id": claims.UserID}).Decode(&storedToken)
    if err != nil || storedToken.AccessToken != tokenString || storedToken.IsValidToken == "n" {
        return nil, errors.New("Token not found or expired")
    }
    return claims, nil
}

func JWTAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusOK, gin.H{
                "code":  http.StatusUnauthorized,
                "error": "Authorization token not provided",
            })
            c.Abort()
            return
        }
        
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusOK, gin.H{
                "code":  http.StatusUnauthorized,
                "error": "Invalid token format",
            })
            c.Abort()
            return
        }
        
        token := parts[1]
        claims, err := ValidateToken(token)
        if err != nil {
            c.JSON(http.StatusOK, gin.H{
                "code":  http.StatusUnauthorized,
                "error": "Invalid or expired token",
            })
            c.Abort()
            return
        }
        
        userID, err := strconv.ParseUint(claims.UserID, 10, 32)
        if err != nil {
            c.JSON(http.StatusOK, gin.H{
                "code":  http.StatusUnauthorized,
                "error": "Invalid user ID format",
            })
            c.Abort()
            return
        }
        
        c.Set("userID", claims.UserID)
        c.Set("user_id", uint(userID))
        c.Set("outlet_id", claims.OutletID)
        
        var user model.User
        if err := database.DbCore.Where("id = ?", uint(userID)).First(&user).Error; err == nil {
            c.Set("username", user.NamaLengkap)
        }
        
        c.Next()
    }
}