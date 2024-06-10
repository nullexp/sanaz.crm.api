package utility

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/misc"
	"github.com/stretchr/testify/assert"
)

func TestJwtToken(t *testing.T) {
	t.Run("Must create and get valid token", func(t *testing.T) {
		got, err := CreateToken(misc.Subject{UserId: "1", SubAccess: "10"}, time.Now().Add(10*time.Minute))

		assert.Equal(t, nil, err)

		jw, err := GetToken(got)
		assert.Equal(t, nil, err)
		assert.NotEqual(t, nil, jw.Audience)
	})

	t.Run("Must return error on invalid signature", func(t *testing.T) {
		got, err := CreateToken(misc.Subject{UserId: "1", SubAccess: "10"}, time.Now().Add(10*time.Minute))
		got = got + "a"
		assert.Equal(t, nil, err)

		_, err = GetToken(got)
		assert.NotEqual(t, nil, err)
	})

	t.Run("Check token: must verify ok", func(t *testing.T) {
		got, err := CreateToken(misc.Subject{UserId: "1", SubAccess: "10"}, time.Now().Add(10*time.Minute))

		assert.Equal(t, nil, err)

		ok, err := CheckToken(got)
		assert.Equal(t, nil, err)
		assert.Equal(t, true, ok)
	})

	t.Run("Check token: must return false on bad signature", func(t *testing.T) {
		data, _ := json.Marshal(misc.Subject{UserId: "1", SubAccess: "10"})

		j := jwt.New(jwt.GetSigningMethod("HS256"))
		j.Claims = misc.StandardClaims{Subject: string(data), ExpiresAt: time.Now().Unix(), Identity: uuid.NewString()}
		got, err := j.SignedString([]byte{15, 18, 25, 68, 18, 93, 25, 95, 14, 89})
		assert.Equal(t, nil, err)

		ok, err := CheckToken(got)
		assert.Equal(t, nil, err)
		assert.Equal(t, false, ok)
	})
}
