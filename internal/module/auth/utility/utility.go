package utility

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/misc"
)

var (
	Salt        []byte
	TokenSecret = []byte("$C&F)J@NcRfUjXnZr4u7x!A%D*G-KaPd")
)

const (
	SaltSize int = 32
	HashSize int = 64
)

func ToSubject(subject string) (out misc.Subject, err error) {
	data, err := base64.RawStdEncoding.DecodeString(subject)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(data, &out)
	return
}

func MustGetSubject(subject string) misc.Subject {
	out, err := ToSubject(subject)
	if err != nil {
		panic(err)
	}
	return out
}

func NewSubject(accessMode string, userId string, subAccess string) misc.Subject {
	return misc.Subject{AccessMode: accessMode, UserId: userId, SubAccess: subAccess}
}

func CreateToken(sb misc.Subject, expire time.Time) (string, error) {
	data, err := json.Marshal(sb)
	if err != nil {
		return "", err
	}

	enc := base64.RawStdEncoding.EncodeToString(data)
	return CreateTokenWithText(enc, expire)
}

func CreateTokenWithText(sb string, expire time.Time) (string, error) {
	t := jwt.New(jwt.GetSigningMethod("HS256"))
	t.Claims = misc.StandardClaims{Subject: sb, ExpiresAt: expire.Unix(), Identity: uuid.NewString()}
	return t.SignedString(TokenSecret)
}

func GetToken(tokenString string) (misc.StandardClaims, error) {
	sc := misc.StandardClaims{}

	rawToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return TokenSecret, nil
	})
	if err != nil {
		return sc, err
	}

	parts := strings.Split(rawToken.Raw, ".")
	if len(parts) != 3 {
		return sc, errors.New("Unknown claim")
	}
	data, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return sc, err
	}
	err = json.Unmarshal(data, &sc)
	if err != nil {
		return sc, errors.New("Unknown subject")
	}

	return sc, nil
}

func CheckToken(tokenString string) (bool, error) {
	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return false, nil
		}
		return TokenSecret, nil
	})

	return err == nil, nil
}
