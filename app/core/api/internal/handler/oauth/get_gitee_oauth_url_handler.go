package oauth

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"schisandra-album-cloud-microservices/app/core/api/internal/logic/oauth"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
)

func GetGiteeOauthUrlHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := oauth.NewGetGiteeOauthUrlLogic(r.Context(), svcCtx)
		resp, err := l.GetGiteeOauthUrl()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
