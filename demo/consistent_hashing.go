package demo

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"

	"google.golang.org/grpc/balancer"
)

type Node struct {
	Name string
	Conn balancer.SubConn
}

type ConsistentHash struct {
	mu             *sync.RWMutex
	hashFunc       func(data []byte) uint32
	virtualNodes   int             // 每个物理节点的虚拟节点数
	sortedHashKeys []uint32        // 排序后的虚拟节点哈希值
	nodeMap        map[uint32]Node // 虚拟节点到物理节点映射
	physicalNodes  map[string]Node // 物理节点集合
}

func NewConsistentHash(virtualNodes int) *ConsistentHash {
	return &ConsistentHash{
		mu:            &sync.RWMutex{},
		hashFunc:      crc32.ChecksumIEEE,
		virtualNodes:  virtualNodes,
		nodeMap:       make(map[uint32]Node),
		physicalNodes: make(map[string]Node),
	}
}

// 添加物理节点
func (c *ConsistentHash) Add(node Node) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.add(node)
}

func (c *ConsistentHash) add(node Node) {
	if _, exists := c.physicalNodes[node.Name]; exists {
		c.physicalNodes[node.Name] = node
		return
	}

	// 生成虚拟节点
	for i := 0; i < c.virtualNodes; i++ {
		hashNode := node.Name + strconv.Itoa(i)
		virtualKey := c.hashFunc([]byte(hashNode))
		c.nodeMap[virtualKey] = node
		c.sortedHashKeys = append(c.sortedHashKeys, virtualKey)
	}

	c.physicalNodes[node.Name] = node
	sort.Slice(c.sortedHashKeys, func(i, j int) bool {
		return c.sortedHashKeys[i] < c.sortedHashKeys[j]
	})
}

// 删除物理节点
func (c *ConsistentHash) Remove(node Node) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.remove(node)
}

func (c *ConsistentHash) remove(node Node) {
	if _, exists := c.physicalNodes[node.Name]; !exists {
		return
	}
	// 删除所有虚拟节点
	newKeys := make([]uint32, 0, len(c.sortedHashKeys))
	for _, key := range c.sortedHashKeys {
		if c.nodeMap[key] != node {
			newKeys = append(newKeys, key)
		} else {
			delete(c.nodeMap, key)
		}
	}

	c.sortedHashKeys = newKeys
	delete(c.physicalNodes, node.Name)
}

// 查找Key对应的节点
func (c *ConsistentHash) Get(key string) Node {
	if len(c.sortedHashKeys) == 0 {
		return Node{}
	}

	hash := c.hashFunc([]byte(key))
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 二分查找最近节点
	idx := sort.Search(len(c.sortedHashKeys), func(i int) bool {
		return c.sortedHashKeys[i] >= hash
	})

	if idx >= len(c.sortedHashKeys) {
		idx = 0
	}

	return c.nodeMap[c.sortedHashKeys[idx]]
}

// 全量替换
func (c *ConsistentHash) AllReplace(nodes map[string]Node) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 删除节点
	for key := range c.physicalNodes {
		physicalNode := c.physicalNodes[key]
		if _, ok := nodes[physicalNode.Name]; !ok {
			c.remove(physicalNode)
		}
	}
	// 添加节点
	for key := range nodes {
		node := nodes[key]
		c.add(node)
	}
}
