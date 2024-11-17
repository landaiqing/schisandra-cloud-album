package websocket

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lxzan/gws"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/common/jwt"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
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
	redis *redis.Client
	ctx   context.Context
	gws.BuiltinEventHandler
	sessions *gws.ConcurrentMap[string, *gws.Conn] // 使用内置的ConcurrentMap存储连接, 可以减少锁冲突
}

var MessageWebSocketHandler *MessageWebSocket

// InitializeWebSocketHandler 初始化WebSocketHandler
func InitializeWebSocketHandler(svcCtx *svc.ServiceContext) {
	MessageWebSocketHandler = NewMessageWebSocket(svcCtx)
}
func (l *MessageWebsocketLogic) MessageWebsocket(w http.ResponseWriter, r *http.Request) {
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
			token := r.URL.Query().Get("token")
			if token == "" {
				return false
			}
			accessToken, res := jwt.ParseAccessToken(l.svcCtx.Config.Auth.AccessSecret, token)
			if !res || accessToken.UserID != clientId {
				return false
			}
			session.Store("user_id", clientId)
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

// NewMessageWebSocket 创建WebSocket实例
func NewMessageWebSocket(svcCtx *svc.ServiceContext) *MessageWebSocket {
	return &MessageWebSocket{
		redis:    svcCtx.RedisClient,
		ctx:      context.Background(),
		sessions: gws.NewConcurrentMap[string, *gws.Conn](64, 128),
	}
}

// OnOpen 连接建立
func (c *MessageWebSocket) OnOpen(socket *gws.Conn) {
	clientId := MustLoad[string](socket.Session(), "user_id")
	c.sessions.Store(clientId, socket)
	// 订阅该用户的频道
	go c.subscribeUserChannel(clientId, c.redis)
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
}

// SendMessageToClient 向指定客户端发送消息
func (c *MessageWebSocket) SendMessageToClient(clientId string, message []byte) error {
	conn, ok := c.sessions.Load(clientId)
	if ok {
		return conn.WriteMessage(gws.OpcodeText, message)
	}
	return fmt.Errorf("client %s not found", clientId)
}

// SendMessageToUser 发送消息到指定用户的 Redis 频道
func (c *MessageWebSocket) SendMessageToUser(clientId string, message []byte, redis *redis.Client, ctx context.Context) error {
	if _, ok := c.sessions.Load(clientId); ok {
		return redis.Publish(ctx, clientId, message).Err()
	} else {
		return redis.LPush(ctx, constant.CommentOfflineMessagePrefix+clientId, message).Err()
	}
}

// 订阅用户频道
func (c *MessageWebSocket) subscribeUserChannel(clientId string, redis *redis.Client) {
	conn, ok := c.sessions.Load(clientId)
	if !ok {
		return
	}

	// 获取离线消息
	messages, err := redis.LRange(c.ctx, constant.CommentOfflineMessagePrefix+clientId, 0, -1).Result()
	if err != nil {
		fmt.Printf("Error loading offline messages for user %s: %v\n", clientId, err)
		return
	}

	// 逐条发送离线消息
	for _, msg := range messages {
		if writeErr := conn.WriteMessage(gws.OpcodeText, []byte(msg)); writeErr != nil {
			fmt.Printf("Error writing offline message to user %s: %v\n", clientId, writeErr)
			return
		}
	}

	// 清空离线消息列表
	if delErr := redis.Del(c.ctx, constant.CommentOfflineMessagePrefix+clientId); delErr.Err() != nil {
		fmt.Printf("Error clearing offline messages for user %s: %v\n", clientId, delErr.Err())
		return
	}

	pubsub := redis.Subscribe(c.ctx, clientId)
	defer func() {
		if closeErr := pubsub.Close(); closeErr != nil {
			fmt.Printf("Error closing pubsub for user %s: %v\n", clientId, closeErr)
		}
	}()

	for {
		msg, waitErr := pubsub.ReceiveMessage(context.Background())
		if waitErr != nil {
			fmt.Printf("Error receiving message for user %s: %v\n", clientId, err)
			return
		}

		if writeErr := conn.WriteMessage(gws.OpcodeText, []byte(msg.Payload)); writeErr != nil {
			fmt.Printf("Error writing message to user %s: %v\n", clientId, writeErr)
			return
		}
	}
}
