package websocket

import (
	"net/http"
	"schisandra-album-cloud-microservices/app/auth/api/internal/logic/websocket"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
)

func FileWebsocketHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := websocket.NewFileWebsocketLogic(r.Context(), svcCtx)
		l.FileWebsocket(w, r)
	}
}
