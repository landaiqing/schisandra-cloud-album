package phone

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"schisandra-album-cloud-microservices/app/auth/api/internal/logic/phone"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
	"schisandra-album-cloud-microservices/common/xhttp"
)

func SharePhoneUploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SharePhoneUploadRequest
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		l := phone.NewSharePhoneUploadLogic(r.Context(), svcCtx)
		err := l.SharePhoneUpload(r, &req)
		xhttp.JsonBaseResponseCtx(r.Context(), w, err)

	}
}
