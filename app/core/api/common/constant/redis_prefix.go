package constant

// 用户相关的redis key
const (
	UserClientPrefix   string = "user:client:"
	UserSessionPrefix  string = "user:session:"
	UserCaptchaPrefix  string = "user:captcha:"
	UserTokenPrefix    string = "user:token:"
	UserSmsRedisPrefix string = "user:sms:"
	UserQrcodePrefix          = "user:qrcode:"
)

// 评论相关的redis key
const (
	CommentSubmitCaptchaPrefix  = "comment:submit:captcha:"
	CommentOfflineMessagePrefix = "comment:offline:message:"
	CommentLikeLockPrefix       = "comment:like:lock:"
	CommentDislikeLockPrefix    = "comment:dislike:lock:"
	CommentLikeListPrefix       = "comment:like:list:"
	CommentUserListPrefix       = "comment:user:list:"
)

// 系统相关的redis key
const (
	SystemApiNoncePrefix = "system:nonce:"
)
