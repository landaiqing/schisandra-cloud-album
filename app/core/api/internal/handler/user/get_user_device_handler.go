package user

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/internal/logic/user"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
)

func GetUserDeviceHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := user.NewGetUserDeviceLogic(r.Context(), svcCtx)
		err := l.GetUserDevice(r)
		if err != nil {
			logx.Error(err)
			httpx.WriteJsonCtx(
				r.Context(),
				w,
				http.StatusInternalServerError,
				response.ErrorWithI18n(r.Context(), "system.error"))
		} else {
			httpx.Ok(w)
		}
	}
}
