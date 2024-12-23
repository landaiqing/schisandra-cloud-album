package websocket

import (
	"context"
	"fmt"
	"net/http"
	"schisandra-album-cloud-microservices/common/jwt"
	"time"

	"github.com/lxzan/gws"
	"github.com/zeromicro/go-zero/core/logx"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
)

type MessageWebsocketLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMessageWebsocketLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MessageWebsocketLogic {
	return &MessageWebsocketLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type MessageWebSocket struct {
	ctx context.Context
	gws.BuiltinEventHandler
	sessions *gws.ConcurrentMap[string, *gws.Conn] // 使用内置的ConcurrentMap存储连接, 可以减少锁冲突
}

var MessageWebSocketHandler = NewMessageWebSocket()

func (l *MessageWebsocketLogic) MessageWebsocket(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Sec-Websocket-Protocol")
	accessToken, res := jwt.ParseAccessToken(l.svcCtx.Config.Auth.AccessSecret, token)
	if !res {
		return
	}
	upgrader := gws.NewUpgrader(MessageWebSocketHandler, &gws.ServerOption{
		HandshakeTimeout: 5 * time.Second, // 握手超时时间
		ReadBufferSize:   1024,            // 读缓冲区大小
		ParallelEnabled:  true,            // 开启并行消息处理
		Recovery:         gws.Recovery,    // 开启异常恢复
		CheckUtf8Enabled: false,           // 关闭UTF8校验
		PermessageDeflate: gws.PermessageDeflate{
			Enabled: true, // 开启压缩
		},
		Authorize: func(r *http.Request, session gws.SessionStorage) bool {
			var clientId = r.URL.Query().Get("user_id")
			if clientId == "" {
				return false
			}
			if accessToken.UserID != clientId {
				return false
			}
			//token := r.URL.Query().Get("token")
			//if token == "" {
			//	return false
			//}
			//accessToken, res := jwt.ParseAccessToken(l.svcCtx.Config.Auth.AccessSecret, token)
			//if !res || accessToken.UserID != clientId {
			//	return false
			//}
			session.Store("user_id", clientId)
			return true
		},
		SubProtocols: []string{token},
	})
	socket, err := upgrader.Upgrade(w, r)
	if err != nil {
		panic(err)
	}
	go func() {
		socket.ReadLoop()
	}()
}

// NewMessageWebSocket 创建WebSocket实例
func NewMessageWebSocket() *MessageWebSocket {
	return &MessageWebSocket{
		ctx:      context.Background(),
		sessions: gws.NewConcurrentMap[string, *gws.Conn](64, 128),
	}
}

// OnOpen 连接建立
func (c *MessageWebSocket) OnOpen(socket *gws.Conn) {
	clientId := MustLoad[string](socket.Session(), "user_id")
	c.sessions.Store(clientId, socket)
	// 订阅该用户的频道
	fmt.Printf("websocket client %s connected\n", clientId)
}

// OnClose 关闭连接
func (c *MessageWebSocket) OnClose(socket *gws.Conn, err error) {
	name := MustLoad[string](socket.Session(), "user_id")
	sharding := c.sessions.GetSharding(name)
	c.sessions.Delete(name)
	sharding.Lock()
	defer sharding.Unlock()
	fmt.Printf("websocket closed, name=%s, msg=%s\n", name, err.Error())
}

// OnPing 处理客户端的Ping消息
func (c *MessageWebSocket) OnPing(socket *gws.Conn, payload []byte) {
	_ = socket.SetDeadline(time.Now().Add(PingInterval + HeartbeatWaitTimeout))
	_ = socket.WritePong(payload)
}

// OnPong 处理客户端的Pong消息
func (c *MessageWebSocket) OnPong(_ *gws.Conn, _ []byte) {}

// OnMessage 接受消息
func (c *MessageWebSocket) OnMessage(socket *gws.Conn, message *gws.Message) {
	defer message.Close()
	clientId := MustLoad[string](socket.Session(), "user_id")
	if conn, ok := c.sessions.Load(clientId); ok {
		_ = conn.WriteMessage(gws.OpcodeText, message.Bytes())
	}
	// fmt.Printf("received message from client %s\n", message.Data)
}
