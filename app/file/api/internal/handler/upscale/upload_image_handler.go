package upscale

import (
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	"schisandra-album-cloud-microservices/app/file/api/internal/logic/upscale"
	"schisandra-album-cloud-microservices/app/file/api/internal/svc"
	"schisandra-album-cloud-microservices/app/file/api/internal/types"

	"schisandra-album-cloud-microservices/common/xhttp"
)

func UploadImageHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UploadRequest
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		l := upscale.NewUploadImageLogic(r.Context(), svcCtx)
		err := l.UploadImage(r, &req)
		xhttp.JsonBaseResponseCtx(r.Context(), w, err)
	}
}
