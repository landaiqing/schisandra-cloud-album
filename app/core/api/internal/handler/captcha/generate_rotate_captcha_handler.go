package captcha

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"schisandra-album-cloud-microservices/app/core/api/internal/logic/captcha"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
)

func GenerateRotateCaptchaHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := captcha.NewGenerateRotateCaptchaLogic(r.Context(), svcCtx)
		resp, err := l.GenerateRotateCaptcha()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
