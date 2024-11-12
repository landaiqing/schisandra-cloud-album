package user

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"schisandra-album-cloud-microservices/app/auth/internal/logic/user"
	"schisandra-album-cloud-microservices/app/auth/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/internal/types"
)

func PhoneLoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PhoneLoginRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewPhoneLoginLogic(r.Context(), svcCtx)
		resp, err := l.PhoneLogin(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
