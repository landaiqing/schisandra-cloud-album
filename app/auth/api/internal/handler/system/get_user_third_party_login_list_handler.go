package system

import (
	"net/http"

	"schisandra-album-cloud-microservices/app/auth/api/internal/logic/system"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/common/xhttp"
)

func GetUserThirdPartyLoginListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := system.NewGetUserThirdPartyLoginListLogic(r.Context(), svcCtx)
		resp, err := l.GetUserThirdPartyLoginList()
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
