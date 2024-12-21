package types

import "encoding/json"

type RedisToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UID          string `json:"uid"`
	Revoked      bool   `json:"revoked" default:"false"`
	AllowAgent   string `json:"allow_agent"`
	GeneratedAt  string `json:"generated_at"`
	UpdatedAt    string `json:"updated_at"`
	GeneratedIP  string `json:"generated_ip"`
}

func (res RedisToken) MarshalBinary() ([]byte, error) {
	return json.Marshal(res)
}

func (res RedisToken) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &res)
}
