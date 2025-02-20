package constant

// 用户相关的redis key
const (
	UserClientPrefix    string = "user:client:"
	UserCaptchaPrefix   string = "user:captcha:"
	UserTokenPrefix     string = "user:token:"
	UserSmsRedisPrefix  string = "user:sms:"
	UserQrcodePrefix           = "user:qrcode:"
	UserOssConfigPrefix        = "user:oss:"
)

// 系统相关的redis key
const (
	SystemApiNoncePrefix = "system:nonce:"
)

const (
	FaceSamplePrefix = "face:samples:"
	FaceVectorPrefix = "face:vectors:"
)

const (
	ImageListPrefix       = "image:list:"
	ImageRecentPrefix     = "image:recent:"
	ImageFaceListPrefix   = "image:faces:"
	ImageSinglePrefix     = "image:single:"
	ImageSharePrefix      = "image:share:"
	ImageShareVisitPrefix = "image:share:visit:"
)
