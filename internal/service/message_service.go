package service

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"rocket-leaf/internal/model"
	"rocket-leaf/internal/rocketmq"
)

// MessageService 消息查询服务
type MessageService struct {
	nextID int64
}

// NewMessageService 创建消息查询服务
func NewMessageService() *MessageService {
	return &MessageService{
		nextID: 1,
	}
}

func (s *MessageService) getNextID() int {
	return int(atomic.AddInt64(&s.nextID, 1))
}

// QueryMessages 查询消息
func (s *MessageService) QueryMessages(topic string, key string, maxResults int) ([]*model.MessageItem, error) {
	client, err := rocketmq.GetClientManager().GetDefaultClient()
	if err != nil {
		return nil, fmt.Errorf("获取客户端失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if maxResults <= 0 {
		maxResults = 32
	}

	msgs, err := client.QueryMessage(ctx, topic, key, maxResults, 0, time.Now().UnixMilli())
	if err != nil {
		return nil, fmt.Errorf("查询消息失败: %w", err)
	}

	result := make([]*model.MessageItem, 0, len(msgs))
	for _, msg := range msgs {
		// MessageExt 使用 Properties map 获取 Tags 和 Keys
		tags := ""
		keys := ""
		if msg.Properties != nil {
			if t, ok := msg.Properties["TAGS"]; ok {
				tags = t
			}
			if k, ok := msg.Properties["KEYS"]; ok {
				keys = k
			}
		}

		item := &model.MessageItem{
			ID:             s.getNextID(),
			Topic:          msg.Topic,
			MessageID:      msg.MsgId,
			Tags:           tags,
			Keys:           keys,
			QueueID:        msg.QueueId,
			QueueOffset:    msg.QueueOffset,
			StoreHost:      msg.StoreHost,
			BornHost:       msg.BornHost,
			StoreTime:      time.Unix(msg.StoreTimestamp/1000, 0).Format("2006-01-02 15:04:05"),
			StoreTimestamp: msg.StoreTimestamp,
			Body:           string(msg.Body),
			Properties:     msg.Properties,
			Status:         model.MsgNormal,
		}

		result = append(result, item)
	}

	return result, nil
}

// QueryMessageByID 按消息 ID 查询消息
func (s *MessageService) QueryMessageByID(topic string, msgID string) (*model.MessageItem, error) {
	client, err := rocketmq.GetClientManager().GetDefaultClient()
	if err != nil {
		return nil, fmt.Errorf("获取客户端失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	msg, err := client.ViewMessage(ctx, topic, msgID)
	if err != nil {
		return nil, fmt.Errorf("查询消息失败: %w", err)
	}

	tags := ""
	keys := ""
	if msg.Properties != nil {
		if t, ok := msg.Properties["TAGS"]; ok {
			tags = t
		}
		if k, ok := msg.Properties["KEYS"]; ok {
			keys = k
		}
	}

	item := &model.MessageItem{
		ID:             s.getNextID(),
		Topic:          msg.Topic,
		MessageID:      msg.MsgId,
		Tags:           tags,
		Keys:           keys,
		QueueID:        msg.QueueId,
		QueueOffset:    msg.QueueOffset,
		StoreHost:      msg.StoreHost,
		BornHost:       msg.BornHost,
		StoreTime:      time.Unix(msg.StoreTimestamp/1000, 0).Format("2006-01-02 15:04:05"),
		StoreTimestamp: msg.StoreTimestamp,
		Body:           string(msg.Body),
		Properties:     msg.Properties,
		Status:         model.MsgNormal,
	}

	return item, nil
}

// GetMessageDetail 获取消息详情
func (s *MessageService) GetMessageDetail(topic string, msgID string) (*model.MessageItem, error) {
	return s.QueryMessageByID(topic, msgID)
}

// GetMessageTrack 获取消息轨迹
// 注意：rocketmq-admin-go 库当前版本暂不支持 MessageTrackDetail
func (s *MessageService) GetMessageTrack(topic string, msgID string) ([]map[string]interface{}, error) {
	// TODO: 待 rocketmq-admin-go 库支持后实现
	return []map[string]interface{}{}, nil
}

// ResendMessage 重投消息
func (s *MessageService) ResendMessage(consumerGroup string, clientID string, topic string, msgID string) (string, error) {
	client, err := rocketmq.GetClientManager().GetDefaultClient()
	if err != nil {
		return "", fmt.Errorf("获取客户端失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := client.ConsumeMessageDirectly(ctx, consumerGroup, clientID, topic, msgID)
	if err != nil {
		return "", fmt.Errorf("重投消息失败: %w", err)
	}

	return fmt.Sprintf("消息重投结果: %v", result), nil
}
