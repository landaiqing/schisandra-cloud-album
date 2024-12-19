package user

import (
	"github.com/zeromicro/go-zero/core/logx"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/internal/logic/user"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
)

func GetUserDeviceHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserDeviceRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewGetUserDeviceLogic(r.Context(), svcCtx)
		resp, err := l.GetUserDevice(r, w, &req)
		if err != nil {
			logx.Error(err)
			httpx.WriteJsonCtx(
				r.Context(),
				w,
				http.StatusInternalServerError,
				response.ErrorWithI18n(r.Context(), "system.error"))
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
