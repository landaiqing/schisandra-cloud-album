package captcha

import (
	"net/http"

	"schisandra-album-cloud-microservices/app/auth/api/internal/logic/captcha"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/common/xhttp"
)

func GenerateTextCaptchaHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := captcha.NewGenerateTextCaptchaLogic(r.Context(), svcCtx)
		resp, err := l.GenerateTextCaptcha()
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
