package websocket

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"schisandra-album-cloud-microservices/app/core/api/internal/logic/websocket"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
)

func MessageWebsocketHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := websocket.NewMessageWebsocketLogic(r.Context(), svcCtx)
		err := l.MessageWebsocket(w, r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}
