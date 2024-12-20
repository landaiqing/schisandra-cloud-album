package oauth

import (
	"github.com/ArtisanCloud/PowerLibs/v3/http/helper"
	"github.com/zeromicro/go-zero/core/logx"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/internal/logic/oauth"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
)

func WechatOffiaccountCallbackVerifyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := oauth.NewWechatOffiaccountCallbackVerifyLogic(r.Context(), svcCtx)
		res, err := l.WechatOffiaccountCallbackVerify(r)
		if err != nil {
			logx.Error(err)
			httpx.WriteJsonCtx(
				r.Context(),
				w,
				http.StatusInternalServerError,
				response.ErrorWithI18n(r.Context(), "system.error"))
		} else {
			_ = helper.HttpResponseSend(res, w)
		}
	}
}
