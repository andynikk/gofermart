package token

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"net/http"

	"github.com/andynikk/gofermart/internal/constants"
)

type Claims struct {
	Authorized bool
	User       string
	Exp        int64
}

func (c *Claims) GenerateJWT() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = c.Authorized
	claims["user"] = c.User
	claims["exp"] = c.Exp

	tokenString, err := token.SignedString(constants.HashKey)

	if err != nil {
		constants.Logger.ErrorLog(err)
	}

	return tokenString, nil
}

func ExtractClaims(tokenStr string) (jwt.MapClaims, bool) {
	hmacSecret := constants.HashKey
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return hmacSecret, nil
	})
	if err != nil {
		return nil, false
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, true
	} else {
		constants.Logger.InfoLog("Invalid JWT Token")
		return nil, false
	}
}

func IsAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Connection", "close")
		//defer r.Body.Close()

		if r.Header["Authorization"] != nil {
			token, err := jwt.Parse(r.Header["Authorization"][0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("there was an error")
				}
				return constants.HashKey, nil
			})

			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				w.Header().Add("Content-Type", "application/json")
				return
			}

			if token.Valid {
				endpoint(w, r)
			}

		} else {
			w.WriteHeader(http.StatusUnauthorized)
			_, err := w.Write([]byte("Not Authorized"))
			if err != nil {
				constants.Logger.ErrorLog(err)
			}
		}
	})
}
