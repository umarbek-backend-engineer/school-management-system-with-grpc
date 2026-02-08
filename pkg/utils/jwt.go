package utils

import (
	"os"
	"sync"
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

		claims["exp"] = jwt.NewNumericDate(time.Now().Add(duration)).Unix()
	} else {
		claims["exp"] = jwt.NewNumericDate(time.Now().Add(time.Hour * 1)).Unix()
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

var JwtStore = JWTStore{
	Tokens: make(map[string]time.Time),
}

type JWTStore struct {
	Mu     sync.Mutex
	Tokens map[string]time.Time
}

func (s *JWTStore) AddToken(token string, exptime time.Time) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.Tokens[token] = exptime
}

func (s *JWTStore) CleanUpExpiredTokens() {
	for {
		time.Sleep(2 * time.Minute)

		s.Mu.Lock()
		for token, timeStamp := range s.Tokens {
			if time.Now().After(timeStamp) {
				delete(s.Tokens, token)
			}
		}
		s.Mu.Unlock()
	}
}

func(s *JWTStore) IsLoggedOut(token string ) bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	_, ok := s.Tokens[token]

	return ok
}