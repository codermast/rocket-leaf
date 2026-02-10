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

// TopicService Topic 管理服务
type TopicService struct {
	nextID int64
}

// NewTopicService 创建 Topic 管理服务
func NewTopicService() *TopicService {
	return &TopicService{
		nextID: 1,
	}
}

func (s *TopicService) getNextID() int {
	return int(atomic.AddInt64(&s.nextID, 1))
}

// GetTopics 获取所有 Topic 列表
func (s *TopicService) GetTopics() ([]*model.TopicItem, error) {
	client, err := rocketmq.GetClientManager().GetDefaultClient()
	if err != nil {
		// 无连接时返回空列表
		return []*model.TopicItem{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	topicList, err := client.FetchAllTopicList(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取 Topic 列表失败: %w", err)
	}

	result := make([]*model.TopicItem, 0, len(topicList.TopicList))
	for _, topic := range topicList.TopicList {
		if isSystemTopic(topic) {
			continue
		}

		item := &model.TopicItem{
			ID:          s.getNextID(),
			Topic:       topic,
			LastUpdated: formatNow(),
		}

		result = append(result, item)
	}

	return result, nil
}

// GetTopicTotal 获取 Topic 总数（排除系统 Topic）
func (s *TopicService) GetTopicTotal() (int, error) {
	client, err := rocketmq.GetClientManager().GetDefaultClient()
	if err != nil {
		// 无连接时返回 0
		return 0, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	topicList, err := client.FetchAllTopicList(ctx)
	if err != nil {
		return 0, fmt.Errorf("获取 Topic 总数失败: %w", err)
	}

	total := 0
	for _, topic := range topicList.TopicList {
		if isSystemTopic(topic) {
			continue
		}
		total++
	}

	return total, nil
}

// GetTopicsByCluster 按集群获取 Topic 列表
func (s *TopicService) GetTopicsByCluster(clusterName string) ([]*model.TopicItem, error) {
	client, err := rocketmq.GetClientManager().GetDefaultClient()
	if err != nil {
		return nil, fmt.Errorf("获取客户端失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	topicList, err := client.FetchTopicsByCluster(ctx, clusterName)
	if err != nil {
		return nil, fmt.Errorf("获取集群 Topic 列表失败: %w", err)
	}

	result := make([]*model.TopicItem, 0, len(topicList.TopicList))
	for _, topic := range topicList.TopicList {
		if isSystemTopic(topic) {
			continue
		}

		item := &model.TopicItem{
			ID:          s.getNextID(),
			Topic:       topic,
			Cluster:     clusterName,
			LastUpdated: formatNow(),
		}

		result = append(result, item)
	}

	return result, nil
}

// GetTopicDetail 获取 Topic 详情
func (s *TopicService) GetTopicDetail(topicName string) (*model.TopicItem, error) {
	client, err := rocketmq.GetClientManager().GetDefaultClient()
	if err != nil {
		return nil, fmt.Errorf("获取客户端失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	routeInfo, err := client.ExamineTopicRouteInfo(ctx, topicName)
	if err != nil {
		return nil, fmt.Errorf("获取 Topic 路由信息失败: %w", err)
	}

	item := &model.TopicItem{
		ID:          s.getNextID(),
		Topic:       topicName,
		Routes:      make([]model.TopicRouteItem, 0),
		LastUpdated: formatNow(),
	}

	totalReadQueue := 0
	totalWriteQueue := 0

	for _, queueData := range routeInfo.QueueDatas {
		route := model.TopicRouteItem{
			Broker:     queueData.BrokerName,
			ReadQueue:  queueData.ReadQueueNums,
			WriteQueue: queueData.WriteQueueNums,
			Perm:       model.IntToPerm(queueData.Perm),
		}

		// BrokerAddrs 是 map[string]string，key 是 "0" 表示 master
		for _, brokerData := range routeInfo.BrokerDatas {
			if brokerData.BrokerName == queueData.BrokerName {
				if addr, ok := brokerData.BrokerAddrs["0"]; ok {
					route.BrokerAddr = addr
				}
				if item.Cluster == "" {
					item.Cluster = brokerData.Cluster
				}
				break
			}
		}

		item.Routes = append(item.Routes, route)
		totalReadQueue += queueData.ReadQueueNums
		totalWriteQueue += queueData.WriteQueueNums
	}

	item.ReadQueue = totalReadQueue
	item.WriteQueue = totalWriteQueue
	if len(item.Routes) > 0 {
		item.Perm = item.Routes[0].Perm
	}

	return item, nil
}

// GetTopicRoute 获取 Topic 路由信息
func (s *TopicService) GetTopicRoute(topicName string) ([]model.TopicRouteItem, error) {
	detail, err := s.GetTopicDetail(topicName)
	if err != nil {
		return nil, err
	}
	return detail.Routes, nil
}

// CreateTopic 创建 Topic
func (s *TopicService) CreateTopic(topic string, brokerAddr string, readQueue int, writeQueue int, perm string) error {
	client, err := rocketmq.GetClientManager().GetDefaultClient()
	if err != nil {
		return fmt.Errorf("获取客户端失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 使用 CreateTopic 方法
	config := admin.TopicConfig{
		TopicName:      topic,
		ReadQueueNums:  readQueue,
		WriteQueueNums: writeQueue,
		Perm:           model.PermToInt(model.TopicPerm(perm)),
	}

	err = client.CreateTopic(ctx, brokerAddr, config)
	if err != nil {
		return fmt.Errorf("创建 Topic 失败: %w", err)
	}

	return nil
}

// UpdateTopic 更新 Topic 配置
func (s *TopicService) UpdateTopic(topic string, brokerAddr string, readQueue int, writeQueue int, perm string) error {
	return s.CreateTopic(topic, brokerAddr, readQueue, writeQueue, perm)
}

// DeleteTopic 删除 Topic
func (s *TopicService) DeleteTopic(topic string, clusterName string) error {
	client, err := rocketmq.GetClientManager().GetDefaultClient()
	if err != nil {
		return fmt.Errorf("获取客户端失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.DeleteTopic(ctx, topic, clusterName)
	if err != nil {
		return fmt.Errorf("删除 Topic 失败: %w", err)
	}

	return nil
}

// GetTopicStats 获取 Topic 统计信息
func (s *TopicService) GetTopicStats(topic string) (map[string]interface{}, error) {
	client, err := rocketmq.GetClientManager().GetDefaultClient()
	if err != nil {
		return nil, fmt.Errorf("获取客户端失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stats, err := client.ExamineTopicStats(ctx, topic)
	if err != nil {
		return nil, fmt.Errorf("获取 Topic 统计失败: %w", err)
	}

	result := map[string]interface{}{
		"topic": topic,
		"stats": stats,
	}

	return result, nil
}

// 判断是否为系统 Topic
func isSystemTopic(topic string) bool {
	systemTopics := []string{
		"SCHEDULE_TOPIC_XXXX",
		"RMQ_SYS_TRANS_HALF_TOPIC",
		"RMQ_SYS_TRACE_TOPIC",
		"RMQ_SYS_TRANS_OP_HALF_TOPIC",
		"TRANS_CHECK_MAX_TIME_TOPIC",
		"SELF_TEST_TOPIC",
		"TBW102",
		"BenchmarkTest",
		"DefaultCluster",
		"OFFSET_MOVED_EVENT",
	}

	for _, st := range systemTopics {
		if topic == st {
			return true
		}
	}

	if len(topic) > 0 && topic[0] == '%' {
		return true
	}

	return false
}
