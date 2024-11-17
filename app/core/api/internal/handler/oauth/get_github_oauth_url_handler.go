package oauth

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	"schisandra-album-cloud-microservices/app/core/api/internal/logic/oauth"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
)

func GetGithubOauthUrlHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OAuthRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := oauth.NewGetGithubOauthUrlLogic(r.Context(), svcCtx)
		resp, err := l.GetGithubOauthUrl(&req)
		if err != nil {
			logx.Error(err)
			httpx.WriteJsonCtx(r.Context(), w, http.StatusInternalServerError, resp)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
