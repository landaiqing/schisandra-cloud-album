package comment

import (
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	"schisandra-album-cloud-microservices/app/auth/api/internal/logic/comment"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"schisandra-album-cloud-microservices/common/xhttp"
)

func GetReplyListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReplyListRequest
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		l := comment.NewGetReplyListLogic(r.Context(), svcCtx)
		resp, err := l.GetReplyList(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
