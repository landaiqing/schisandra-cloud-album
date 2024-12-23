package captcha

import (
	"net/http"
	"schisandra-album-cloud-microservices/common/xhttp"

	"schisandra-album-cloud-microservices/app/auth/api/internal/logic/captcha"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
)

func GenerateRotateCaptchaHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := captcha.NewGenerateRotateCaptchaLogic(r.Context(), svcCtx)
		resp, err := l.GenerateRotateCaptcha()
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
