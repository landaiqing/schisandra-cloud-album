package client

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"schisandra-album-cloud-microservices/app/core/api/common/utils"
	"schisandra-album-cloud-microservices/app/core/api/internal/logic/client"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
)

func GenerateClientIdHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := utils.GetClientIP(r)
		l := client.NewGenerateClientIdLogic(r.Context(), svcCtx)
		resp, err := l.GenerateClientId(clientIP)
		if err != nil || resp.Code == 500 {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
