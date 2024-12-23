package oauth

import (
	"net/http"
	"schisandra-album-cloud-microservices/common/xhttp"

	"schisandra-album-cloud-microservices/app/auth/api/internal/logic/oauth"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
)

func GetGiteeOauthUrlHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := oauth.NewGetGiteeOauthUrlLogic(r.Context(), svcCtx)
		resp, err := l.GetGiteeOauthUrl()
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
