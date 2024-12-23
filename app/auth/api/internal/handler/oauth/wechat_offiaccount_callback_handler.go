package oauth

import (
	"github.com/ArtisanCloud/PowerLibs/v3/http/helper"
	"net/http"
	"schisandra-album-cloud-microservices/app/auth/api/internal/logic/oauth"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/common/xhttp"
)

func WechatOffiaccountCallbackHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := oauth.NewWechatOffiaccountCallbackLogic(r.Context(), svcCtx)
		res, err := l.WechatOffiaccountCallback(r)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			_ = helper.HttpResponseSend(res, w)
		}
	}
}
