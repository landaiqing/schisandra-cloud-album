package oauth

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"schisandra-album-cloud-microservices/app/core/api/internal/logic/oauth"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
)

func WechatCallbackHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := oauth.NewWechatCallbackLogic(r.Context(), svcCtx)
		err := l.WechatCallback(w, r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}
