package types

import "encoding/json"

// SessionData 返回数据
type SessionData struct {
	RefreshToken string `json:"refresh_token"`
	UID          string `json:"uid"`
}

func (res SessionData) MarshalBinary() ([]byte, error) {
	return json.Marshal(res)
}

func (res SessionData) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &res)
}
