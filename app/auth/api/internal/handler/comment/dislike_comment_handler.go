package comment

import (
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	"schisandra-album-cloud-microservices/app/auth/api/internal/logic/comment"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"schisandra-album-cloud-microservices/common/xhttp"
)

func DislikeCommentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CommentDisLikeRequest
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		l := comment.NewDislikeCommentLogic(r.Context(), svcCtx)
		err := l.DislikeComment(&req)
		xhttp.JsonBaseResponseCtx(r.Context(), w, err)
	}
}
