package types

import "encoding/json"

type RedisToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UID          string `json:"uid"`
	Revoked      bool   `json:"revoked" default:"false"`
}

func (res RedisToken) MarshalBinary() ([]byte, error) {
	return json.Marshal(res)
}

func (res RedisToken) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &res)
}
