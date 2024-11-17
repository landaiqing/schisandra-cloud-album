package websocket

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lxzan/gws"
	"github.com/zeromicro/go-zero/core/logx"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/common/utils"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
)

type QrcodeWebsocketLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQrcodeWebsocketLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QrcodeWebsocketLogic {
	return &QrcodeWebsocketLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

const (
	PingInterval         = 5 * time.Second  // 客户端心跳间隔
	HeartbeatWaitTimeout = 10 * time.Second // 心跳等待超时时间
)

type QrcodeWebSocket struct {
	gws.BuiltinEventHandler
	sessions *gws.ConcurrentMap[string, *gws.Conn] // 使用内置的ConcurrentMap存储连接, 可以减少锁冲突
}

var QrcodeWebSocketHandler = NewWebSocket()

func (l *QrcodeWebsocketLogic) QrcodeWebsocket(w http.ResponseWriter, r *http.Request) {
	upgrader := gws.NewUpgrader(QrcodeWebSocketHandler, &gws.ServerOption{
		HandshakeTimeout: 5 * time.Second, // 握手超时时间
		ReadBufferSize:   1024,            // 读缓冲区大小
		ParallelEnabled:  true,            // 开启并行消息处理
		Recovery:         gws.Recovery,    // 开启异常恢复
		CheckUtf8Enabled: false,           // 关闭UTF8校验
		PermessageDeflate: gws.PermessageDeflate{
			Enabled: true, // 开启压缩
		},
		Authorize: func(r *http.Request, session gws.SessionStorage) bool {
			var clientId = r.URL.Query().Get("client_id")
			if clientId == "" {
				return false
			}
			ip := utils.GetClientIP(r)
			exists := l.svcCtx.RedisClient.Get(l.ctx, constant.UserClientPrefix+ip).Val()
			if clientId != exists {
				return false
			}
			session.Store("client_id", clientId)
			return true
		},
	})
	socket, err := upgrader.Upgrade(w, r)
	if err != nil {
		panic(err)
	}
	go func() {
		socket.ReadLoop() // 此处阻塞会使请求上下文不能顺利被GC
	}()
}

// MustLoad 从session中加载数据
func MustLoad[T any](session gws.SessionStorage, key string) (v T) {
	if value, exist := session.Load(key); exist {
		v = value.(T)
	}
	return
}

// NewWebSocket 创建WebSocket实例
func NewWebSocket() *QrcodeWebSocket {
	return &QrcodeWebSocket{
		sessions: gws.NewConcurrentMap[string, *gws.Conn](64, 128),
	}
}

// OnOpen 连接建立
func (c *QrcodeWebSocket) OnOpen(socket *gws.Conn) {
	clientId := MustLoad[string](socket.Session(), "client_id")
	c.sessions.Store(clientId, socket)
	fmt.Printf("websocket client %s connected\n", clientId)
}

// OnClose 关闭连接
func (c *QrcodeWebSocket) OnClose(socket *gws.Conn, err error) {
	name := MustLoad[string](socket.Session(), "client_id")
	sharding := c.sessions.GetSharding(name)
	c.sessions.Delete(name)
	sharding.Lock()
	defer sharding.Unlock()
	fmt.Printf("websocket closed, name=%s, err=%s\n", name, err.Error())
}

// OnPing 处理客户端的Ping消息
func (c *QrcodeWebSocket) OnPing(socket *gws.Conn, payload []byte) {
	_ = socket.SetDeadline(time.Now().Add(PingInterval + HeartbeatWaitTimeout))
	_ = socket.WritePong(payload)
}

// OnPong 处理客户端的Pong消息
func (c *QrcodeWebSocket) OnPong(_ *gws.Conn, _ []byte) {}

// OnMessage 接受消息
func (c *QrcodeWebSocket) OnMessage(socket *gws.Conn, message *gws.Message) {
	defer message.Close()
	clientId := MustLoad[string](socket.Session(), "client_id")
	if conn, ok := c.sessions.Load(clientId); ok {
		_ = conn.WriteMessage(gws.OpcodeText, message.Bytes())
	}
}

// SendMessageToClient 向指定客户端发送消息
func (c *QrcodeWebSocket) SendMessageToClient(clientId string, message []byte) error {
	conn, ok := c.sessions.Load(clientId)
	if ok {
		return conn.WriteMessage(gws.OpcodeText, message)
	}
	return fmt.Errorf("client %s not found", clientId)
}
