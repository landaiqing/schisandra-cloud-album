package oauth

import (
	"errors"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	"schisandra-album-cloud-microservices/app/core/api/internal/logic/oauth"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
)

func GiteeCallbackHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OAuthCallbackRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := oauth.NewGiteeCallbackLogic(r.Context(), svcCtx)
		err := l.GiteeCallback(w, r, &req)
		if err != nil {
			logx.Error(err)
			httpx.ErrorCtx(r.Context(), w, errors.New("server error"))
		} else {
			httpx.Ok(w)
		}
	}
}
