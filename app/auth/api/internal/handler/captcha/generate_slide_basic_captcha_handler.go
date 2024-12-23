package captcha

import (
	"net/http"
	"schisandra-album-cloud-microservices/common/xhttp"

	"schisandra-album-cloud-microservices/app/auth/api/internal/logic/captcha"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
)

func GenerateSlideBasicCaptchaHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := captcha.NewGenerateSlideBasicCaptchaLogic(r.Context(), svcCtx)
		resp, err := l.GenerateSlideBasicCaptcha()
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
