package comment

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	"schisandra-album-cloud-microservices/app/core/api/internal/logic/comment"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
)

func GetReplyListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReplyListRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := comment.NewGetReplyListLogic(r.Context(), svcCtx)
		resp, err := l.GetReplyList(&req)
		if err != nil {
			logx.Error(err)
			httpx.WriteJsonCtx(r.Context(), w, http.StatusInternalServerError, resp)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
