package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"rocket-leaf/internal/model"
	"rocket-leaf/internal/rocketmq"
)

// ClusterService 集群状态服务
type ClusterService struct {
	connectionService *ConnectionService
}

// NewClusterService 创建集群状态服务
func NewClusterService(connService *ConnectionService) *ClusterService {
	return &ClusterService{
		connectionService: connService,
	}
}

// GetClusterInfo 获取集群信息
func (s *ClusterService) GetClusterInfo() (*model.ClusterInfo, error) {
	client, err := rocketmq.GetClientManager().GetDefaultClient()
	if err != nil {
		// 无连接时返回空数据
		return &model.ClusterInfo{
			Brokers:     make([]*model.BrokerNode, 0),
			NameServers: make([]string, 0),
		}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 获取集群信息
	clusterInfo, err := client.ExamineBrokerClusterInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取集群信息失败: %w", err)
	}

	result := &model.ClusterInfo{
		NameServers: client.GetNameServerAddressList(),
		Brokers:     make([]*model.BrokerNode, 0),
	}

	// 统计集群信息
	brokerID := 1
	for clusterName, brokerNames := range clusterInfo.ClusterAddrTable {
		result.ClusterName = clusterName

		for _, brokerName := range brokerNames {
			brokerData, exists := clusterInfo.BrokerAddrTable[brokerName]
			if !exists {
				continue
			}

			// BrokerAddrs 是 map[string]string，key 是 brokerId 的字符串形式
			for brokerIDStr, addr := range brokerData.BrokerAddrs {
				role := model.RoleSlave
				if brokerIDStr == "0" {
					role = model.RoleMaster
				}

				brokerIDInt, _ := strconv.Atoi(brokerIDStr)
				broker := &model.BrokerNode{
					ID:         brokerID,
					Cluster:    clusterName,
					BrokerName: brokerName,
					BrokerID:   brokerIDInt,
					Role:       role,
					Address:    addr,
					Status:     model.NodeOnline,
					LastUpdate: formatNow(),
				}

				result.Brokers = append(result.Brokers, broker)
				brokerID++
			}
		}
	}

	result.TotalBrokers = len(result.Brokers)
	result.OnlineBrokers = len(result.Brokers)

	return result, nil
}

// GetBrokers 获取 Broker 列表
func (s *ClusterService) GetBrokers() ([]*model.BrokerNode, error) {
	clusterInfo, err := s.GetClusterInfo()
	if err != nil {
		return []*model.BrokerNode{}, nil
	}
	return clusterInfo.Brokers, nil
}

// GetBrokerDetail 获取 Broker 详情
func (s *ClusterService) GetBrokerDetail(brokerAddr string) (*model.BrokerNode, error) {
	client, err := rocketmq.GetClientManager().GetDefaultClient()
	if err != nil {
		return nil, fmt.Errorf("获取客户端失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stats, err := client.FetchBrokerRuntimeStats(ctx, brokerAddr)
	if err != nil {
		return nil, fmt.Errorf("获取 Broker 统计信息失败: %w", err)
	}

	broker := &model.BrokerNode{
		Address:    brokerAddr,
		Status:     model.NodeOnline,
		LastUpdate: formatNow(),
	}

	if stats != nil && stats.Table != nil {
		if version, ok := stats.Table["brokerVersionDesc"]; ok {
			broker.Version = version
		}
		if tpsIn, ok := stats.Table["putTps"]; ok {
			broker.TpsIn = parseIntSafe(extractFirstValue(tpsIn))
		}
		if tpsOut, ok := stats.Table["getTransferredTps"]; ok {
			broker.TpsOut = parseIntSafe(extractFirstValue(tpsOut))
		}
		if msgInToday, ok := stats.Table["msgPutTotalTodayNow"]; ok {
			broker.MsgInToday = parseInt64Safe(msgInToday)
		}
		if msgOutToday, ok := stats.Table["msgGetTotalTodayNow"]; ok {
			broker.MsgOutToday = parseInt64Safe(msgOutToday)
		}
		if diskRatio, ok := stats.Table["commitLogDiskRatio"]; ok {
			broker.CommitLogDiskUsage = int(parseFloatSafe(diskRatio) * 100)
		}
	}

	return broker, nil
}

// GetNameServers 获取 NameServer 列表
func (s *ClusterService) GetNameServers() ([]*model.NameServerNode, error) {
	client, err := rocketmq.GetClientManager().GetDefaultClient()
	if err != nil {
		// 无连接时返回空数据
		return []*model.NameServerNode{}, nil
	}

	addrs := client.GetNameServerAddressList()

	result := make([]*model.NameServerNode, 0, len(addrs))
	for i, addr := range addrs {
		node := &model.NameServerNode{
			ID:       i + 1,
			Address:  addr,
			Status:   model.NodeOnline,
			LastSeen: formatNow(),
		}
		result = append(result, node)
	}

	return result, nil
}

// GetClusterSummary 获取集群概览统计
func (s *ClusterService) GetClusterSummary() (*model.ClusterSummary, error) {
	clusterInfo, err := s.GetClusterInfo()
	if err != nil {
		return &model.ClusterSummary{}, nil
	}

	summary := &model.ClusterSummary{
		TotalClusters:  1,
		TotalBrokers:   clusterInfo.TotalBrokers,
		OnlineBrokers:  clusterInfo.OnlineBrokers,
		WarningBrokers: 0,
		OfflineBrokers: 0,
		AvgDiskUsage:   clusterInfo.AvgDiskUsage,
	}

	return summary, nil
}

// RefreshBrokerStats 刷新 Broker 统计信息
func (s *ClusterService) RefreshBrokerStats(brokerAddr string) (*model.BrokerNode, error) {
	return s.GetBrokerDetail(brokerAddr)
}

// 辅助函数
func parseIntSafe(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

func parseInt64Safe(s string) int64 {
	var result int64
	fmt.Sscanf(s, "%d", &result)
	return result
}

func parseFloatSafe(s string) float64 {
	var result float64
	fmt.Sscanf(s, "%f", &result)
	return result
}

func extractFirstValue(s string) string {
	for i, c := range s {
		if c == ' ' || c == '\t' {
			return s[:i]
		}
	}
	return s
}
