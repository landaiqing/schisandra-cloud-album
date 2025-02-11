package storage

import (
	"net/http"

	"schisandra-album-cloud-microservices/app/auth/api/internal/logic/storage"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/common/xhttp"
)

func QueryLocationImageListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := storage.NewQueryLocationImageListLogic(r.Context(), svcCtx)
		resp, err := l.QueryLocationImageList()
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
