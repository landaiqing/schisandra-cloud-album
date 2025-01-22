package constant

// 用户相关的redis key
const (
	UserClientPrefix   string = "user:client:"
	UserCaptchaPrefix  string = "user:captcha:"
	UserTokenPrefix    string = "user:token:"
	UserSmsRedisPrefix string = "user:sms:"
	UserQrcodePrefix          = "user:qrcode:"
)

// 系统相关的redis key
const (
	SystemApiNoncePrefix = "system:nonce:"
)

const (
	FaceSamplePrefix = "face:sample:"
)
