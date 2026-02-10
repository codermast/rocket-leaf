// Package service 提供业务服务层
package service

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"rocket-leaf/internal/model"
	"rocket-leaf/internal/rocketmq"
)

// ConnectionService 连接管理服务
type ConnectionService struct {
	mu          sync.RWMutex
	connections map[int]*model.Connection // 连接配置列表
	nextID      int                       // 下一个连接ID
}

// NewConnectionService 创建连接管理服务
func NewConnectionService() *ConnectionService {
	return &ConnectionService{
		connections: make(map[int]*model.Connection),
		nextID:      1,
	}
}

// formatNow 格式化当前时间
func formatNow() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// GetConnections 获取所有连接配置
func (s *ConnectionService) GetConnections() []*model.Connection {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*model.Connection, 0, len(s.connections))
	for _, conn := range s.connections {
		result = append(result, conn)
	}
	return result
}

// GetConnection 获取单个连接配置
func (s *ConnectionService) GetConnection(id int) (*model.Connection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conn, exists := s.connections[id]
	if !exists {
		return nil, fmt.Errorf("连接不存在: %d", id)
	}
	return conn, nil
}

// AddConnection 添加新连接
func (s *ConnectionService) AddConnection(name string, env string, nameServer string, timeoutSec int, enableACL bool, accessKey string, secretKey string, remark string) (*model.Connection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 验证环境类型
	connEnv := model.ConnectionEnv(env)
	if connEnv != model.EnvProduction && connEnv != model.EnvTest && connEnv != model.EnvDevelopment {
		connEnv = model.EnvDevelopment
	}

	accessKey = strings.TrimSpace(accessKey)
	secretKey = strings.TrimSpace(secretKey)

	if enableACL {
		if strings.TrimSpace(accessKey) == "" {
			return nil, fmt.Errorf("启用 ACL 时 AccessKey 不能为空")
		}
		if strings.TrimSpace(secretKey) == "" {
			return nil, fmt.Errorf("启用 ACL 时 SecretKey 不能为空")
		}
	} else {
		accessKey = ""
		secretKey = ""
	}

	conn := &model.Connection{
		ID:         s.nextID,
		Name:       name,
		Env:        connEnv,
		NameServer: nameServer,
		TimeoutSec: timeoutSec,
		EnableACL:  enableACL,
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		Status:     model.StatusOffline,
		LastCheck:  "-",
		IsDefault:  len(s.connections) == 0, // 第一个连接自动设为默认
		Remark:     remark,
	}

	s.connections[s.nextID] = conn
	s.nextID++

	return conn, nil
}

// UpdateConnection 更新连接配置
func (s *ConnectionService) UpdateConnection(id int, name string, env string, nameServer string, timeoutSec int, enableACL bool, accessKey string, secretKey string, remark string) (*model.Connection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	conn, exists := s.connections[id]
	if !exists {
		return nil, fmt.Errorf("连接不存在: %d", id)
	}

	// 如果 NameServer 地址变更，需要移除旧客户端
	if conn.NameServer != nameServer {
		rocketmq.GetClientManager().RemoveClient(conn.NameServer)
	}

	// 验证环境类型
	connEnv := model.ConnectionEnv(env)
	if connEnv != model.EnvProduction && connEnv != model.EnvTest && connEnv != model.EnvDevelopment {
		connEnv = model.EnvDevelopment
	}

	accessKey = strings.TrimSpace(accessKey)
	secretKey = strings.TrimSpace(secretKey)

	if enableACL {
		if strings.TrimSpace(accessKey) == "" {
			return nil, fmt.Errorf("启用 ACL 时 AccessKey 不能为空")
		}
		if strings.TrimSpace(secretKey) == "" {
			return nil, fmt.Errorf("启用 ACL 时 SecretKey 不能为空")
		}
	} else {
		accessKey = ""
		secretKey = ""
	}

	conn.Name = name
	conn.Env = connEnv
	conn.NameServer = nameServer
	conn.TimeoutSec = timeoutSec
	conn.EnableACL = enableACL
	conn.AccessKey = accessKey
	conn.SecretKey = secretKey
	conn.Remark = remark

	return conn, nil
}

// DeleteConnection 删除连接
func (s *ConnectionService) DeleteConnection(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	conn, exists := s.connections[id]
	if !exists {
		return fmt.Errorf("连接不存在: %d", id)
	}

	// 不允许删除默认连接（如果还有其他连接）
	if conn.IsDefault && len(s.connections) > 1 {
		return fmt.Errorf("不能删除默认连接，请先设置其他连接为默认")
	}

	// 移除客户端
	rocketmq.GetClientManager().RemoveClient(conn.NameServer)

	delete(s.connections, id)

	// 如果删除后还有连接，设置第一个为默认
	if len(s.connections) > 0 && conn.IsDefault {
		for _, c := range s.connections {
			c.IsDefault = true
			break
		}
	}

	return nil
}

// TestConnection 测试连接
func (s *ConnectionService) TestConnection(id int) (string, error) {
	s.mu.Lock()
	conn, exists := s.connections[id]
	if !exists {
		s.mu.Unlock()
		return "", fmt.Errorf("连接不存在: %d", id)
	}
	nameServer := conn.NameServer
	timeout := time.Duration(conn.TimeoutSec) * time.Second
	enableACL := conn.EnableACL
	accessKey := conn.AccessKey
	secretKey := conn.SecretKey
	s.mu.Unlock()

	// 测试连接
	err := rocketmq.GetClientManager().TestConnection(nameServer, timeout, enableACL, accessKey, secretKey)

	s.mu.Lock()
	defer s.mu.Unlock()

	// 更新连接状态
	if conn, exists := s.connections[id]; exists {
		conn.LastCheck = formatNow()
		if err == nil {
			conn.Status = model.StatusOnline
			return "online", nil
		} else {
			conn.Status = model.StatusOffline
			return "offline", err
		}
	}

	return "offline", err
}

// SetDefaultConnection 设置默认连接
func (s *ConnectionService) SetDefaultConnection(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	conn, exists := s.connections[id]
	if !exists {
		return fmt.Errorf("连接不存在: %d", id)
	}

	// 取消其他连接的默认状态
	for _, c := range s.connections {
		c.IsDefault = false
	}

	conn.IsDefault = true

	// 更新客户端管理器的默认连接
	return rocketmq.GetClientManager().SetDefaultConnection(conn.NameServer)
}

// ConnectDefault 连接默认连接
func (s *ConnectionService) ConnectDefault() error {
	s.mu.RLock()
	var defaultConn *model.Connection
	for _, conn := range s.connections {
		if conn.IsDefault {
			defaultConn = conn
			break
		}
	}
	s.mu.RUnlock()

	if defaultConn == nil {
		return fmt.Errorf("无默认连接配置")
	}

	timeout := time.Duration(defaultConn.TimeoutSec) * time.Second
	_, err := rocketmq.GetClientManager().CreateClient(
		defaultConn.NameServer,
		timeout,
		defaultConn.EnableACL,
		defaultConn.AccessKey,
		defaultConn.SecretKey,
	)
	if err != nil {
		return err
	}

	return rocketmq.GetClientManager().SetDefaultConnection(defaultConn.NameServer)
}

// Connect 连接指定连接
func (s *ConnectionService) Connect(id int) error {
	s.mu.RLock()
	conn, exists := s.connections[id]
	if !exists {
		s.mu.RUnlock()
		return fmt.Errorf("连接不存在: %d", id)
	}
	nameServer := conn.NameServer
	timeout := time.Duration(conn.TimeoutSec) * time.Second
	enableACL := conn.EnableACL
	accessKey := conn.AccessKey
	secretKey := conn.SecretKey
	s.mu.RUnlock()

	_, err := rocketmq.GetClientManager().CreateClient(nameServer, timeout, enableACL, accessKey, secretKey)
	if err != nil {
		return err
	}

	// 更新连接状态
	s.mu.Lock()
	if conn, exists := s.connections[id]; exists {
		conn.Status = model.StatusOnline
		conn.LastCheck = formatNow()
	}
	s.mu.Unlock()

	return nil
}
