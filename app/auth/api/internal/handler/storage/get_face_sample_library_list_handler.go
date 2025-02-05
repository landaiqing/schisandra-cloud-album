package storage

import (
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"schisandra-album-cloud-microservices/app/auth/api/internal/logic/storage"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/common/xhttp"
)

func GetFaceSampleLibraryListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FaceSampleLibraryListRequest
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		l := storage.NewGetFaceSampleLibraryListLogic(r.Context(), svcCtx)
		resp, err := l.GetFaceSampleLibraryList(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
