package client

import (
	"net/http"
	"schisandra-album-cloud-microservices/common/utils"
	"schisandra-album-cloud-microservices/common/xhttp"

	"schisandra-album-cloud-microservices/app/auth/api/internal/logic/client"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
)

func GenerateClientIdHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := utils.GetClientIP(r)
		l := client.NewGenerateClientIdLogic(r.Context(), svcCtx)
		resp, err := l.GenerateClientId(clientIP)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
