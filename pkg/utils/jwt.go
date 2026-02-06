package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func SingingJWT(id, username, role string) (string, error) {

	// loading secret key from .env file
	jwtSercretKey := os.Getenv("JWT_SECRETE_STRING")
	jwtExpIn := os.Getenv("JWT_EXPIRES_IN")

	// building claims for jwt token
	claims := jwt.MapClaims{
		"uid":      id,
		"username": username,
		"role":     role,
	}

	// checking if the secrete key is not empty so that jwt will no be set with empty key
	if jwtSercretKey == "" {
		return "", ErrorHandler(nil, "JWT secret key is missing")
	}

	// set token expiration so expired tokens cannot be used
	if jwtExpIn != "" {
		duration, err := time.ParseDuration(jwtExpIn)
		if err != nil {
			return "", ErrorHandler(err, "Internal error")
		}

		claims["exp"] = jwt.NewNumericDate(time.Now().Add(duration))
	} else {
		claims["exp"] = jwt.NewNumericDate(time.Now().Add(time.Hour * 1))
	}

	// making token with all the claims + defining the signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// signing the token with the secret key
	signedToken, err := token.SignedString([]byte(jwtSercretKey))
	if err != nil {
		return "", ErrorHandler(err, "Internal error")
	}

	return signedToken, nil
}
