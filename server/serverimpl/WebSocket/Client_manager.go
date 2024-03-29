package WebSocket

import (
	"FantasticLife/utils"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"sync"
	"time"
)

const (
	DefaultAppId = 101 // 默认平台Id
)

type DisposeFunc func(client *Client, seq string, message []byte) (code uint32, msg string, data interface{})

// 连接管理
type ClientManager struct {
	Clients         map[*Client]bool   // 全部的连接
	ClientsLock     sync.RWMutex       // 读写锁
	Users           map[string]*Client // 登录的用户 // appId+uuid
	UserLock        sync.RWMutex       // 读写锁
	RegisterChan    chan *Client       // 连接连接处理
	Login           chan *login        // 用户登录处理
	Unregister      chan *Client       // 断开连接处理程序
	Broadcast       chan []byte        // 广播 向全部成员发送数据
	Handlers        map[string]DisposeFunc
	HandlersRWMutex sync.RWMutex
	appIds          []uint32
	logger          *zap.Logger
	RedisCli        *redis.Client
	MysqlCli        *gorm.DB
}

func NewClientManager(logger *zap.Logger, RedisCli *redis.Client, MysqlCli *gorm.DB) (clientManager *ClientManager) {
	clientManager = &ClientManager{
		Clients:         make(map[*Client]bool),
		Users:           make(map[string]*Client),
		RegisterChan:    make(chan *Client, 1000),
		Login:           make(chan *login, 1000),
		Unregister:      make(chan *Client, 1000),
		Broadcast:       make(chan []byte, 1000),
		Handlers:        make(map[string]DisposeFunc),
		HandlersRWMutex: sync.RWMutex{},
		appIds:          []uint32{DefaultAppId, 102, 103, 104},
		logger:          logger,
		RedisCli:        RedisCli,
		MysqlCli:        MysqlCli,
	}
	clientManager.Register("login", LoginController)
	clientManager.Register("heartbeat", HeartbeatController)
	clientManager.Register("ping", PingController)

	return
}

// 连接功能
// handler注册到全局的map handlers[key]里
func (manager *ClientManager) Register(key string, value DisposeFunc) {
	manager.HandlersRWMutex.Lock()
	defer manager.HandlersRWMutex.Unlock()
	manager.Handlers[key] = value
	return
}

func (manager *ClientManager) GetHandlers(key string) (value DisposeFunc, ok bool) {
	manager.HandlersRWMutex.RLock()
	defer manager.HandlersRWMutex.RUnlock()

	value, ok = manager.Handlers[key]

	return
}

// 获取用户key
func GetUserKey(appId uint32, userId string) (key string) {
	key = fmt.Sprintf("%d_%s", appId, userId)

	return
}

/**************************  manager  ***************************************/

func (manager *ClientManager) InClient(client *Client) (ok bool) {
	manager.ClientsLock.RLock()
	defer manager.ClientsLock.RUnlock()

	// 连接存在，在添加
	_, ok = manager.Clients[client]

	return
}

// GetClients
func (manager *ClientManager) GetClients() (clients map[*Client]bool) {

	clients = make(map[*Client]bool)

	manager.ClientsRange(func(client *Client, value bool) (result bool) {
		clients[client] = value

		return true
	})

	return
}

// 遍历
func (manager *ClientManager) ClientsRange(f func(client *Client, value bool) (result bool)) {

	manager.ClientsLock.RLock()
	defer manager.ClientsLock.RUnlock()

	for key, value := range manager.Clients {
		result := f(key, value)
		if result == false {
			return
		}
	}

	return
}

// GetClientsLen
func (manager *ClientManager) GetClientsLen() (clientsLen int) {

	clientsLen = len(manager.Clients)

	return
}

// 添加客户端
func (manager *ClientManager) AddClients(client *Client) {
	manager.ClientsLock.Lock()
	defer manager.ClientsLock.Unlock()

	manager.Clients[client] = true
}

// 删除客户端
func (manager *ClientManager) DelClients(client *Client) {
	manager.ClientsLock.Lock()
	defer manager.ClientsLock.Unlock()

	if _, ok := manager.Clients[client]; ok {
		delete(manager.Clients, client)
	}
}

// 获取用户的Client用户连接类
func (manager *ClientManager) GetUserClient(appId uint32, userId string) (client *Client) {

	manager.UserLock.RLock()
	defer manager.UserLock.RUnlock()

	userKey := GetUserKey(appId, userId)
	if value, ok := manager.Users[userKey]; ok {
		client = value
	}

	return
}

// GetClientsLen
func (manager *ClientManager) GetUsersLen() (userLen int) {
	userLen = len(manager.Users)

	return
}

// 添加用户
func (manager *ClientManager) AddUsers(key string, client *Client) {
	manager.UserLock.Lock()
	defer manager.UserLock.Unlock()

	manager.Users[key] = client
}

// 删除用户
func (manager *ClientManager) DelUsers(client *Client) (result bool) {
	manager.UserLock.Lock()
	defer manager.UserLock.Unlock()

	key := GetUserKey(client.AppId, client.UserId)
	if value, ok := manager.Users[key]; ok {
		// 判断是否为相同的用户
		if value.Addr != client.Addr {

			return
		}
		delete(manager.Users, key)
		result = true
	}

	return
}

// 获取用户的key
func (manager *ClientManager) GetUserKeys() (userKeys []string) {

	userKeys = make([]string, 0)
	manager.UserLock.RLock()
	defer manager.UserLock.RUnlock()
	for key := range manager.Users {
		userKeys = append(userKeys, key)
	}

	return
}

// 获取用户的key
func (manager *ClientManager) GetUserList(appId uint32) (userList []string) {

	userList = make([]string, 0)

	manager.UserLock.RLock()
	defer manager.UserLock.RUnlock()

	for _, v := range manager.Users {
		if v.AppId == appId {
			userList = append(userList, v.UserId)
		}
	}

	fmt.Println("GetUserList len:", len(manager.Users))

	return
}

// 获取用户的key
func (manager *ClientManager) GetUserClients() (clients []*Client) {

	clients = make([]*Client, 0)
	manager.UserLock.RLock()
	defer manager.UserLock.RUnlock()
	for _, v := range manager.Users {
		clients = append(clients, v)
	}

	return
}

// 向全部成员(除了自己)发送数据
func (manager *ClientManager) sendAll(message []byte, ignoreClient *Client) {

	clients := manager.GetUserClients()
	for _, conn := range clients {
		if conn != ignoreClient {
			conn.SendMsg(message)
		}
	}
}

// 向全部成员(除了自己)发送数据
func (manager *ClientManager) sendAppIdAll(message []byte, appId uint32, ignoreClient *Client) {

	clients := manager.GetUserClients()
	for _, conn := range clients {
		if conn != ignoreClient && conn.AppId == appId {
			conn.SendMsg(message)
		}
	}
}

// 用户建立连接事件
func (manager *ClientManager) EventRegister(client *Client) {
	manager.AddClients(client)

	manager.logger.Info("EventRegister 用户建立连接", zap.String("Client Addr: ", client.Addr))

	client.Send <- []byte("连接成功")
}

// 用户登录
func (manager *ClientManager) EventLogin(login *login) {

	client := login.Client
	// 连接存在，在添加
	if manager.InClient(client) {
		userKey := login.GetKey()
		manager.AddUsers(userKey, login.Client)
	}

	manager.logger.Info("EventLogin 用户登录", zap.String("addr", client.Addr), zap.Uint32("appId", login.AppId), zap.String("userId", login.UserId))

	orderId := GetOrderIdTime()
	result, err := manager.SendUserMessageAll(login.AppId, login.UserId, orderId, utils.MessageCmdEnter, "哈喽~")
	if err != nil {
		manager.logger.Error("EventLogin 用户登录给全体发消息", zap.Error(err))
		return
	}
	manager.logger.Info("EventLogin 用户登录给全体发消息", zap.Bool("result", result))
}

// 用户断开连接
func (manager *ClientManager) EventUnregister(client *Client) {
	manager.DelClients(client)

	// 删除用户连接
	deleteResult := manager.DelUsers(client)
	if deleteResult == false {
		// 不是当前连接的客户端

		return
	}

	//TODO 清除redis登录数据
	//userOnline, err := cache.GetUserOnlineInfo(client.GetKey())
	//if err == nil {
	//	userOnline.LogOut()
	//	cache.SetUserOnlineInfo(client.GetKey(), userOnline)
	//}

	// 关闭 chan
	// close(client.Send)

	fmt.Println("EventUnregister 用户断开连接", client.Addr, client.AppId, client.UserId)
	manager.logger.Info("EventUnregister 用户断开连接", zap.String("addr", client.Addr), zap.Uint32("appId", client.AppId), zap.String("userId", client.UserId))

	if client.UserId != "" {
		orderId := GetOrderIdTime()
		_, err := manager.SendUserMessageAll(client.AppId, client.UserId, orderId, utils.MessageCmdExit, "用户已经离开~")
		if err != nil {
			manager.logger.Error("EventUnregister 用户断开连接给全体发消息", zap.Error(err))
			return
		}
	}
}

// 管道处理程序，管道事务的处理，包括建立连接、用户登录、断开连接、广播事件
func (manager *ClientManager) start() {
	for {
		select {
		case conn := <-manager.RegisterChan:
			// 建立连接事件
			manager.EventRegister(conn)

		case login := <-manager.Login:
			// 用户登录
			manager.EventLogin(login)

		case conn := <-manager.Unregister:
			// 断开连接事件
			manager.EventUnregister(conn)

		case message := <-manager.Broadcast:
			// 广播事件
			clients := manager.GetClients()
			for conn := range clients {
				select {
				case conn.Send <- message:
				default:
					close(conn.Send)
				}
			}
		}
	}
}

/**************************  manager info  ***************************************/
// 获取管理者信息
func (manager *ClientManager) GetManagerInfo(isDebug string) (managerInfo map[string]interface{}) {
	managerInfo = make(map[string]interface{})

	managerInfo["clientsLen"] = manager.GetClientsLen()        // 客户端连接数
	managerInfo["usersLen"] = manager.GetUsersLen()            // 登录用户数
	managerInfo["chanRegisterLen"] = len(manager.RegisterChan) // 未处理连接事件数
	managerInfo["chanLoginLen"] = len(manager.Login)           // 未处理登录事件数
	managerInfo["chanUnregisterLen"] = len(manager.Unregister) // 未处理退出登录事件数
	managerInfo["chanBroadcastLen"] = len(manager.Broadcast)   // 未处理广播事件数

	if isDebug == "true" {
		addrList := make([]string, 0)
		manager.ClientsRange(func(client *Client, value bool) (result bool) {
			addrList = append(addrList, client.Addr)

			return true
		})

		users := manager.GetUserKeys()

		managerInfo["clients"] = addrList // 客户端列表
		managerInfo["users"] = users      // 登录用户列表
	}

	return
}

// 获取用户所在的连接
func (manager *ClientManager) GetUserClient_bug(appId uint32, userId string) (client *Client) {
	client = manager.GetUserClient(appId, userId)

	return
}

// 定时清理超时连接
func (manager *ClientManager) ClearTimeoutConnections() {
	currentTime := uint64(time.Now().Unix())

	clients := manager.GetClients()
	for client := range clients {
		if client.IsHeartbeatTimeout(currentTime) {
			fmt.Println("心跳时间超时 关闭连接", client.Addr, client.UserId, client.LoginTime, client.HeartbeatTime)

			client.Socket.Close()
		}
	}
}

// 获取全部用户
func (manager *ClientManager) GetUserList_bug(appId uint32) (userList []string) {
	fmt.Println("获取全部用户", appId)

	userList = manager.GetUserList(appId)

	return
}

// 全员广播
func (manager *ClientManager) AllSendMessages(appId uint32, userId string, data string) {
	fmt.Println("全员广播", appId, userId, data)
	//发送消息排除自己
	ignoreClient := manager.GetUserClient(appId, userId)
	manager.sendAppIdAll([]byte(data), appId, ignoreClient)
}

func (manager *ClientManager) InAppIds(appId uint32) bool {
	inAppId := false
	for _, value := range manager.appIds {
		if value == appId {
			inAppId = true

			return inAppId
		}
	}
	return inAppId
}

func (manager *ClientManager) SendUserMessageAll(appId uint32, userId string, msgId, cmd, message string) (sendResults bool, err error) {
	sendResults = true

	//currentTime := uint64(time.Now().Unix())
	//servers, err := cache.GetServerAll(currentTime)
	if err != nil {
		fmt.Println("给全体用户发消息", err)
		manager.logger.Error("给全体用户发消息", zap.Error(err))
		return
	}
	data := utils.GetMsgData(userId, msgId, cmd, message)

	//for _, server := range servers {
	//	if IsLocal(server) {
	//		data := models.GetMsgData(userId, msgId, cmd, message)
	//		AllSendMessages(appId, userId, data)
	//	} else {
	//		grpcclient.SendMsgAll(server, msgId, appId, userId, cmd, message)
	//	}
	//}
	ignoreClient := manager.GetUserClient(appId, userId)
	manager.sendAppIdAll([]byte(data), appId, ignoreClient)
	return sendResults, err
}

func GetOrderIdTime() (orderId string) {

	currentTime := time.Now().Nanosecond()
	orderId = fmt.Sprintf("%d", currentTime)

	return
}
