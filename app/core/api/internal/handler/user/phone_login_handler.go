package user

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	"schisandra-album-cloud-microservices/app/core/api/internal/logic/user"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
)

func PhoneLoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PhoneLoginRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewPhoneLoginLogic(r.Context(), svcCtx)
		resp, err := l.PhoneLogin(r, w, &req)
		if err != nil || resp.Code == 500 {
			logx.Error(err)
			httpx.WriteJsonCtx(r.Context(), w, http.StatusInternalServerError, resp)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
