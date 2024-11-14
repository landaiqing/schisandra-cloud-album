package types

import "encoding/json"

type RedisToken struct {
	AccessToken string `json:"access_token"`
	UID         string `json:"uid"`
}

func (res RedisToken) MarshalBinary() ([]byte, error) {
	return json.Marshal(res)
}

func (res RedisToken) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &res)
}
