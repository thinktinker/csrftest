package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"hash"
)

// HMAC is a wrapper around the crypto/hmac package
// making it easier to user in our code.

type HMAC struct {
	hmac hash.Hash
}

// NewHMAC accepts and returns a new HMAC object
// NOTE: Instantiates a NewHMAC with the secret key

func NewHMAC(key string) HMAC {
	h := hmac.New(sha256.New, []byte(key))
	return HMAC{
		hmac: h,
	}
}

// Hash will hash the provided input string using HMAC with
// the secret key provided using the HMAC object that was created
// NOTE: Hashes the input string (i.e. the Remember Token String)
// with the secret key in the hmac object

func (h HMAC) Hash(input string) string {
	h.hmac.Reset()
	h.hmac.Write([]byte(input))
	b := h.hmac.Sum(nil)
	return base64.URLEncoding.EncodeToString(b) // takes the b byte slice that have values and maps to a string values that is url-safe
}
