package types

// SessionData 返回数据
type SessionData struct {
	RefreshToken string `json:"refresh_token"`
	UID          string `json:"uid"`
}
