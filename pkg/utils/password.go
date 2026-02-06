package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// HashPassword takes a plain-text password, hashes it securely using Argon2,
// and returns a single encoded string containing both the salt and the hash.
// Format of returned string: "saltBase64.hashBase64"
func HashPassword(password string) (string, error) {

	// Validate input: password must not be empty
	if password == "" {
		return "", ErrorHandler(fmt.Errorf("object is empty"), "Password must be filled!")
	}

	// Generate a random 16-byte salt
	// Salt makes identical passwords produce different hashes
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", ErrorHandler(err, "Internal Error while generating salt")
	}

	// Generate Argon2 hash using:
	// - password + salt
	// - time cost = 1 iteration
	// - memory cost = 64MB
	// - parallelism = 8 threads
	// - output hash length = 32 bytes
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 8, 32)

	// Encode salt and hash to Base64 so they can be stored as text in DB
	saltBase64 := base64.StdEncoding.EncodeToString(salt)
	hashBase64 := base64.StdEncoding.EncodeToString(hash)

	// Combine salt and hash into one string separated by "."
	// Example: "salt123.hash456"
	encodedHash := fmt.Sprintf("%s.%s", saltBase64, hashBase64)

	return encodedHash, nil
}

func VerifyPassword(password, encodedPassword string) error {

	// spliting salt and hashedpassword
	part := strings.Split(encodedPassword, ".")
	if len(part) != 2 {
		return ErrorHandler(fmt.Errorf("Invalid Encodedpassword"), "Bad password")
	}

	// asingning them init variables
	saltbased64 := part[0]
	hashbased64 := part[1]

	// decoding salt and hash to compare them with password
	salt, err := base64.StdEncoding.DecodeString(saltbased64)
	if err != nil {
		return ErrorHandler(err, "Invalid Encodedpassword")
	}

	dbhash, err := base64.StdEncoding.DecodeString(hashbased64)
	if err != nil {
		return ErrorHandler(err, "Invalid Encodedpassword")
	}

	// incoding password to bite slice so it is the same with hash which was incoded earlier
	passwordhash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 8, 32)

	// checking if the passwordhash and dbhash are the same
	// extra security
	if len(passwordhash) != len(dbhash) {
		return ErrorHandler(fmt.Errorf("Passwords do not match"), "Incorrect password")
	}

	// this function will return 1 if the given byte slices are same, if not 0
	if subtle.ConstantTimeCompare(passwordhash, dbhash) == 1 {
		return nil
	}
	return ErrorHandler(fmt.Errorf("Passwords do not match"), "Incorrect password")
}
