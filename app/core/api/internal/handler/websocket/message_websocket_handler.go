package websocket

import (
	"net/http"

	"schisandra-album-cloud-microservices/app/core/api/internal/logic/websocket"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
)

func MessageWebsocketHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := websocket.NewMessageWebsocketLogic(r.Context(), svcCtx)
		l.MessageWebsocket(w, r)
	}
}
