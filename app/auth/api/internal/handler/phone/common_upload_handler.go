package phone

import (
	"net/http"

	"schisandra-album-cloud-microservices/app/auth/api/internal/logic/phone"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/common/xhttp"
)

func CommonUploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := phone.NewCommonUploadLogic(r.Context(), svcCtx)
		err := l.CommonUpload()
		xhttp.JsonBaseResponseCtx(r.Context(), w, err)
	}
}
