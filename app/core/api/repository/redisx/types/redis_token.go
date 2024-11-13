package types

type RedisToken struct {
	AccessToken string `json:"access_token"`
	UID         string `json:"uid"`
}
