package comment

import (
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	"schisandra-album-cloud-microservices/app/community/api/internal/logic/comment"
	"schisandra-album-cloud-microservices/app/community/api/internal/svc"
	"schisandra-album-cloud-microservices/app/community/api/internal/types"
	"schisandra-album-cloud-microservices/common/xhttp"
)

func LikeCommentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CommentLikeRequest
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		l := comment.NewLikeCommentLogic(r.Context(), svcCtx)
		err := l.LikeComment(&req)
		xhttp.JsonBaseResponseCtx(r.Context(), w, err)
	}
}
