package get

import (
	"time"
)

type Token struct {
	AccessToken            string    `json:"accessToken"`
	RefreshToken           string    `json:"refreshToken"`
	AccessTokenExpireTime  time.Time `json:"accessTokenExpireTime"`
	RefreshTokenExpireTime time.Time `json:"refreshTokenExpireTime"`
}
