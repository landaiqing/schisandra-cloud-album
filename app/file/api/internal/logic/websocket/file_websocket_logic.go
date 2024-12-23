package websocket

import (
	"context"
	"fmt"
	"github.com/lxzan/gws"
	"net/http"
	"schisandra-album-cloud-microservices/app/file/api/internal/svc"
	"schisandra-album-cloud-microservices/common/jwt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileWebsocketLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFileWebsocketLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileWebsocketLogic {
	return &FileWebsocketLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

const (
	PingInterval         = 5 * time.Second  // 客户端心跳间隔
	HeartbeatWaitTimeout = 10 * time.Second // 心跳等待超时时间
)

type FileWebSocket struct {
	ctx context.Context
	gws.BuiltinEventHandler
	sessions *gws.ConcurrentMap[string, *gws.Conn]
}

var FileWebSocketHandler = NewFileWebSocket()

func (l *FileWebsocketLogic) FileWebsocket(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Sec-Websocket-Protocol")
	accessToken, res := jwt.ParseAccessToken(l.svcCtx.Config.Auth.AccessSecret, token)
	if !res {
		return
	}
	upGrader := gws.NewUpgrader(FileWebSocketHandler, &gws.ServerOption{
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
			session.Store("user_id", clientId)
			return true
		},
		SubProtocols: []string{token},
	})
	socket, err := upGrader.Upgrade(w, r)
	if err != nil {
		panic(err)
	}
	go func() {
		socket.ReadLoop()
	}()
}

func NewFileWebSocket() *FileWebSocket {
	return &FileWebSocket{
		ctx:      context.Background(),
		sessions: gws.NewConcurrentMap[string, *gws.Conn](64, 128),
	}
}

// OnOpen 连接建立
func (c *FileWebSocket) OnOpen(socket *gws.Conn) {
	clientId := MustLoad[string](socket.Session(), "user_id")
	c.sessions.Store(clientId, socket)
	// 订阅该用户的频道
	fmt.Printf("websocket client %s connected\n", clientId)
}

// OnClose 关闭连接
func (c *FileWebSocket) OnClose(socket *gws.Conn, err error) {
	name := MustLoad[string](socket.Session(), "user_id")
	sharding := c.sessions.GetSharding(name)
	c.sessions.Delete(name)
	sharding.Lock()
	defer sharding.Unlock()
	fmt.Printf("websocket closed, name=%s, msg=%s\n", name, err.Error())
}

// OnPing 处理客户端的Ping消息
func (c *FileWebSocket) OnPing(socket *gws.Conn, payload []byte) {
	_ = socket.SetDeadline(time.Now().Add(PingInterval + HeartbeatWaitTimeout))
	_ = socket.WritePong(payload)
}

// OnPong 处理客户端的Pong消息
func (c *FileWebSocket) OnPong(_ *gws.Conn, _ []byte) {}

// OnMessage 接受消息
func (c *FileWebSocket) OnMessage(socket *gws.Conn, message *gws.Message) {
	defer message.Close()
	clientId := MustLoad[string](socket.Session(), "user_id")
	if conn, ok := c.sessions.Load(clientId); ok {
		_ = conn.WriteMessage(gws.OpcodeText, message.Bytes())
	}
	// fmt.Printf("received message from client %s\n", message.Data)
}

// SendMessageToClient 向指定客户端发送消息
func (c *FileWebSocket) SendMessageToClient(clientId string, message []byte) error {
	conn, ok := c.sessions.Load(clientId)
	if ok {
		return conn.WriteMessage(gws.OpcodeText, message)
	}
	return fmt.Errorf("client %s not found", clientId)
}

// MustLoad 从session中加载数据
func MustLoad[T any](session gws.SessionStorage, key string) (v T) {
	if value, exist := session.Load(key); exist {
		v = value.(T)
	}
	return
}
