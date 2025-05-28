package demo

import (
	"fmt"
	"gitee.com/flycash/permission-platform/pkg/ctxx"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)


type ConsistentHashingBalancer struct {
	consistentHash *ConsistentHash
}

func (r *ConsistentHashingBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	ctx := info.Ctx
	bizId,err := ctxx.GetBizID(ctx)
	if err != nil {
		return balancer.PickResult{}, err
	}
	uid,err := ctxx.GetUID(ctx)
	if err != nil {
		return balancer.PickResult{}, err
	}
	key := fmt.Sprintf("%d-%d", bizId, uid)
	node := r.consistentHash.Get(key)
	return balancer.PickResult{
		SubConn: node.Conn,
	}, nil
}


type ConsistentHashingBalancerBuilder struct {
	consistentHash *ConsistentHash
}

func NewConsistentHashingBalancer(virtualNodes int) *ConsistentHashingBalancer {
	return &ConsistentHashingBalancer{
		consistentHash: NewConsistentHash(virtualNodes),
	}
}

func (w *ConsistentHashingBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	nodeMap := make(map[string]Node)
	for sub, subInfo := range info.ReadySCs {
		nodeName := subInfo.Address.Attributes.Value("node").(string)
		nodeMap[nodeName] = Node{
			Name: nodeName,
			Conn: sub,
		}
	}
	w.consistentHash.AllReplace(nodeMap)
	return &ConsistentHashingBalancer{
		consistentHash: w.consistentHash,
	}
}