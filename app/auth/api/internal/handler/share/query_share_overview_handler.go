package share

import (
	"net/http"

	"schisandra-album-cloud-microservices/app/auth/api/internal/logic/share"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/common/xhttp"
)

func QueryShareOverviewHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := share.NewQueryShareOverviewLogic(r.Context(), svcCtx)
		resp, err := l.QueryShareOverview()
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
