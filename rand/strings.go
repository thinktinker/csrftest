package rand

import (
	"crypto/rand"
	"encoding/base64"
)

const RemberTokenBytes = 32

// Bytes generates random bytes, or will return an error if there was one
// It users the crypto/rand package so it is safe to use for remember tokens

func Bytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b) // Generates random bytes based on the length of the []byte slice
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Nbytes returns the number of bytes in the base64 URL encoded string

func NBytes(base64String string) (int, error) {
	b, err := base64.URLEncoding.DecodeString(base64String)
	if err != nil {
		return -1, err
	}

	return len(b), nil
}

// String generates a byte slice of nBytes and
// returns a string that is based on the base64 URL encoded version
// of the byte slice

func String(nBytes int) (string, error) {
	b, err := Bytes(nBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// RememberToken is a helper function designed to generate
// remember tokens of a predetermined byte size.

func RememberToken() (string, error) {
	return String(RemberTokenBytes)
}
