package auth

import (
	"net/http"

	"schisandra-album-cloud-microservices/app/auth/api/internal/logic/auth"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/common/xhttp"
)

func CheckUserSecuritySettingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := auth.NewCheckUserSecuritySettingLogic(r.Context(), svcCtx)
		resp, err := l.CheckUserSecuritySetting()
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
