package rand

import (
	"crypto/rand"
	"encoding/base64"
)

const RememberTokenBytes = 32

func RememberToken() (string, error) {
	return GenString(RememberTokenBytes)
}

func GenString(nBytes int) (string, error) {
	b, err := GenBytes(nBytes)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

func GenBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
