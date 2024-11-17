package captcha

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/internal/logic/captcha"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
)

func GenerateRotateCaptchaHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := captcha.NewGenerateRotateCaptchaLogic(r.Context(), svcCtx)
		resp, err := l.GenerateRotateCaptcha()
		if err != nil {
			logx.Error(err)
			httpx.WriteJsonCtx(
				r.Context(),
				w,
				http.StatusInternalServerError,
				response.ErrorWithI18n(r.Context(), "system.error"))
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
