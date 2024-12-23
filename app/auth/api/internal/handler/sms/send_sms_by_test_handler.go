package sms

import (
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	"schisandra-album-cloud-microservices/common/xhttp"

	"schisandra-album-cloud-microservices/app/auth/api/internal/logic/sms"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
)

func SendSmsByTestHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SmsSendRequest
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		l := sms.NewSendSmsByTestLogic(r.Context(), svcCtx)
		err := l.SendSmsByTest(&req)
		xhttp.JsonBaseResponseCtx(r.Context(), w, err)
	}
}
