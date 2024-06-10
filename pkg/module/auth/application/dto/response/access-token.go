package get

import (
	"time"
)

type AccessToken struct {
	AccessToken           string    `json:"accessToken"`
	AccessTokenExpireTime time.Time `json:"accessTokenExpireTime"`
}
