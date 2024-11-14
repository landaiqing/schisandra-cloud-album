package user

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"schisandra-album-cloud-microservices/app/core/api/internal/logic/user"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
)

func GetUserDeviceHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := user.NewGetUserDeviceLogic(r.Context(), svcCtx)
		err := l.GetUserDevice(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}
