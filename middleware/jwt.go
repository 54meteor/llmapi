package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/model"
	"strings"
	"time"
)

var jwtSecret = []byte(config.SystemName + "_jwt_secret_key")

type Claims struct {
	UserId   int    `json:"user_id"`
	Username string `json:"username"`
	Role     int    `json:"role"`
	jwt.MapClaims
}

func GenerateToken(user *model.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserId:   user.Id,
		Username: user.Username,
		Role:     user.Role,
		MapClaims: jwt.MapClaims{
			"exp": expirationTime.Unix(),
			"iat": time.Now().Unix(),
			"iss": config.SystemName,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "未提供认证 token",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无效或过期的 token",
			})
			c.Abort()
			return
		}

		user, err := model.GetUserById(claims.UserId, false)
		if err != nil || user.Status == common.UserStatusDisabled {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "用户不存在或已被封禁",
			})
			c.Abort()
			return
		}

		c.Set("id", claims.UserId)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func AuthWithRole(minRole int) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetInt("role")
		if role < minRole {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无权进行此操作，权限不足",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
