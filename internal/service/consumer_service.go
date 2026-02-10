package service

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"rocket-leaf/internal/model"
	"rocket-leaf/internal/rocketmq"

	admin "github.com/codermast/rocketmq-admin-go"
)

// ConsumerService 消费者组服务
type ConsumerService struct {
	nextID int64
}

// NewConsumerService 创建消费者组服务
func NewConsumerService() *ConsumerService {
	return &ConsumerService{
		nextID: 1,
	}
}

func (s *ConsumerService) getNextID() int {
	return int(atomic.AddInt64(&s.nextID, 1))
}

// GetConsumerGroups 获取所有消费者组列表
func (s *ConsumerService) GetConsumerGroups() ([]*model.ConsumerGroupItem, error) {
	client, err := rocketmq.GetClientManager().GetDefaultClient()
	if err != nil {
		// 无连接时返回空列表
		return []*model.ConsumerGroupItem{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clusterInfo, err := client.ExamineBrokerClusterInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取集群信息失败: %w", err)
	}

	result := make([]*model.ConsumerGroupItem, 0)
	processedGroups := make(map[string]bool)

	for _, brokerData := range clusterInfo.BrokerAddrTable {
		// BrokerAddrs 是 map[string]string
		masterAddr, ok := brokerData.BrokerAddrs["0"]
		if !ok {
			continue
		}

		// GetAllSubscriptionGroup 返回 map[string]*SubscriptionGroupConfig
		subGroups, err := client.GetAllSubscriptionGroup(ctx, masterAddr)
		if err != nil {
			continue
		}

		if subGroups == nil {
			continue
		}

		for groupName, config := range subGroups {
			if isSystemGroup(groupName) {
				continue
			}

			if processedGroups[groupName] {
				continue
			}
			processedGroups[groupName] = true

			item := &model.ConsumerGroupItem{
				ID:          s.getNextID(),
				Group:       groupName,
				ConsumeMode: model.ModeClustering,
				Status:      model.GroupOffline,
				MaxRetry:    config.RetryMaxTimes,
				LastUpdate:  formatNow(),
			}

			connInfo, err := client.ExamineConsumerConnectionInfo(ctx, groupName)
			if err == nil && connInfo != nil {
				item.OnlineClients = len(connInfo.ConnectionSet)
				if item.OnlineClients > 0 {
					item.Status = model.GroupOnline
				}

				for _, conn := range connInfo.ConnectionSet {
					c := model.GroupClient{
						ClientID:      conn.ClientId,
						IP:            conn.ClientAddr,
						Version:       fmt.Sprintf("%d", conn.Version),
						LastHeartbeat: formatNow(),
					}
					item.Clients = append(item.Clients, c)
				}

				for topic, expr := range connInfo.SubscriptionTable {
					sub := model.GroupSubscription{
						Topic:      topic,
						Expression: expr.SubString,
					}
					item.Subscriptions = append(item.Subscriptions, sub)
				}
				item.TopicCount = len(item.Subscriptions)
			}

			result = append(result, item)
		}
	}

	return result, nil
}

// GetConsumerGroupDetail 获取消费者组详情
func (s *ConsumerService) GetConsumerGroupDetail(groupName string) (*model.ConsumerGroupItem, error) {
	client, err := rocketmq.GetClientManager().GetDefaultClient()
	if err != nil {
		return nil, fmt.Errorf("获取客户端失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	item := &model.ConsumerGroupItem{
		ID:            s.getNextID(),
		Group:         groupName,
		ConsumeMode:   model.ModeClustering,
		Status:        model.GroupOffline,
		Subscriptions: make([]model.GroupSubscription, 0),
		Clients:       make([]model.GroupClient, 0),
		LastUpdate:    formatNow(),
	}

	connInfo, err := client.ExamineConsumerConnectionInfo(ctx, groupName)
	if err == nil && connInfo != nil {
		item.OnlineClients = len(connInfo.ConnectionSet)
		if item.OnlineClients > 0 {
			item.Status = model.GroupOnline
		}

		if connInfo.ConsumeType == "CONSUME_ACTIVELY" {
			item.ConsumeMode = model.ModeBroadcasting
		}

		for _, conn := range connInfo.ConnectionSet {
			c := model.GroupClient{
				ClientID:      conn.ClientId,
				IP:            conn.ClientAddr,
				Version:       fmt.Sprintf("%d", conn.Version),
				LastHeartbeat: formatNow(),
			}
			item.Clients = append(item.Clients, c)
		}

		for topic, expr := range connInfo.SubscriptionTable {
			sub := model.GroupSubscription{
				Topic:      topic,
				Expression: expr.SubString,
			}
			item.Subscriptions = append(item.Subscriptions, sub)
		}
		item.TopicCount = len(item.Subscriptions)
	}

	return item, nil
}

// GetConsumeStats 获取消费统计信息
func (s *ConsumerService) GetConsumeStats(groupName string) (map[string]interface{}, error) {
	client, err := rocketmq.GetClientManager().GetDefaultClient()
	if err != nil {
		return nil, fmt.Errorf("获取客户端失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ExamineConsumeStats 只有一个参数
	stats, err := client.ExamineConsumeStats(ctx, groupName)
	if err != nil {
		return nil, fmt.Errorf("获取消费统计失败: %w", err)
	}

	// 计算总延迟
	var totalDiff int64
	for _, offset := range stats.OffsetTable {
		totalDiff += offset.BrokerOffset - offset.ConsumerOffset
	}

	result := map[string]interface{}{
		"group":      groupName,
		"consumeTps": stats.ConsumeTps,
		"diffTotal":  totalDiff,
	}

	return result, nil
}

// CreateConsumerGroup 创建消费者组
func (s *ConsumerService) CreateConsumerGroup(group string, brokerAddr string, consumeMode string, maxRetry int) error {
	client, err := rocketmq.GetClientManager().GetDefaultClient()
	if err != nil {
		return fmt.Errorf("获取客户端失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 使用 CreateSubscriptionGroup
	config := admin.SubscriptionGroupConfig{
		GroupName:              group,
		ConsumeEnable:          true,
		ConsumeFromMinEnable:   true,
		ConsumeBroadcastEnable: consumeMode == string(model.ModeBroadcasting),
		RetryMaxTimes:          maxRetry,
	}

	err = client.CreateSubscriptionGroup(ctx, brokerAddr, config)
	if err != nil {
		return fmt.Errorf("创建消费者组失败: %w", err)
	}

	return nil
}

// DeleteConsumerGroup 删除消费者组
func (s *ConsumerService) DeleteConsumerGroup(group string, brokerAddr string) error {
	client, err := rocketmq.GetClientManager().GetDefaultClient()
	if err != nil {
		return fmt.Errorf("获取客户端失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.DeleteSubscriptionGroup(ctx, brokerAddr, group)
	if err != nil {
		return fmt.Errorf("删除消费者组失败: %w", err)
	}

	return nil
}

// ResetOffset 重置消费位点
func (s *ConsumerService) ResetOffset(group string, topic string, timestamp int64, force bool) error {
	client, err := rocketmq.GetClientManager().GetDefaultClient()
	if err != nil {
		return fmt.Errorf("获取客户端失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err = client.ResetOffsetByTimestamp(ctx, topic, group, timestamp, force)
	if err != nil {
		return fmt.Errorf("重置消费位点失败: %w", err)
	}

	return nil
}

// GetConsumerClients 获取消费者客户端列表
func (s *ConsumerService) GetConsumerClients(groupName string) ([]model.GroupClient, error) {
	detail, err := s.GetConsumerGroupDetail(groupName)
	if err != nil {
		return nil, err
	}
	return detail.Clients, nil
}

// 判断是否为系统消费者组
func isSystemGroup(group string) bool {
	systemGroups := []string{
		"CID_ONSAPI_OWNER",
		"CID_ONSAPI_PERMISSION",
		"CID_ONSAPI_PULL",
		"CID_RMQ_SYS_TRANS",
		"TOOLS_CONSUMER",
		"FILTERSRV_CONSUMER",
		"__MONITOR_CONSUMER",
		"CLIENT_INNER_PRODUCER",
		"SELF_TEST_C_GROUP",
		"SELF_TEST_P_GROUP",
		"CID_RMQ_SYS_TRACE",
	}

	for _, sg := range systemGroups {
		if group == sg {
			return true
		}
	}

	if len(group) > 10 && group[:10] == "CID_ONSAPI" {
		return true
	}

	return false
}
