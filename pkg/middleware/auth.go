package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"graduation-thesis/pkg/custom_error"
	request "graduation-thesis/pkg/requests"
)

func extractTokenFromRequest(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}

	return ""
}

func validateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(_token *jwt.Token) (interface{}, error) {
		if _, ok := _token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", _token.Header["alg"])
		}

		return []byte(viper.GetString("app.access_secret")), nil
	})
	if err != nil {
		return nil, err
	}

	if _, ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
		return nil, fmt.Errorf("token has expired")
	}

	return token, nil
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractTokenFromRequest(c.Request)
		token, err := validateToken(tokenString)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, _ := token.Claims.(jwt.MapClaims)
		accessUuid := claims["access_uuid"]
		userId := claims["user_id"]

		c.Set("access_uuid", accessUuid)
		c.Set("user_id", userId)
		c.Next()
	}
}

func AuthMiddlewareV2(authenticatorURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractTokenFromRequest(c.Request)
		var (
			result interface{}
			err    error
		)
		for i := 1; i <= 5; i++ {
			result, err = request.HTTPRequestCall(
				authenticatorURL,
				http.MethodPost,
				tokenString,
				nil,
				5*time.Second,
			)
			if err != nil && !errors.Is(err, custom_error.ErrNoPermission) {
				continue
			}
			break
		}

		if errors.Is(err, custom_error.ErrNoPermission) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		} else if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		userID := result.(string)
		c.Request.Header.Add("X-User-ID", userID)
		c.Next()
	}
}
