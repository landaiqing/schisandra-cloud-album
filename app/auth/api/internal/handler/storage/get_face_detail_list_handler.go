package storage

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"schisandra-album-cloud-microservices/app/auth/api/internal/logic/storage"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
	"schisandra-album-cloud-microservices/common/xhttp"
)

func GetFaceDetailListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FaceDetailListRequest
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		l := storage.NewGetFaceDetailListLogic(r.Context(), svcCtx)
		resp, err := l.GetFaceDetailList(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
