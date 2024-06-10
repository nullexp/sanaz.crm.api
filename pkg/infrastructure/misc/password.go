package misc

import (
	"crypto/sha256"
	b64 "encoding/base64"
)

type Password interface {
	HashAndSalt(password string) string
	ComparePasswords(hashedPwd, plainPwd string) bool
}

type sha256Password struct {
	secret string
}

func NewSha256Password(secret string) Password {
	return sha256Password{secret: secret}
}

func (s sha256Password) HashAndSalt(password string) string {

	combination := s.secret + string(password) + s.secret
	sha256 := sha256.Sum256([]byte(combination))
	return b64.StdEncoding.EncodeToString(sha256[:])
}

func (s sha256Password) ComparePasswords(hashedPwd, plainPwd string) bool {

	byteHash := s.HashAndSalt(plainPwd)

	return hashedPwd == byteHash
}
