//go:build unit
package demo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)



func TestConsistentHash_AllReplace(t *testing.T) {
	ch := NewConsistentHash(3)

	// 初始节点
	initialNodes := map[string]Node{
		"node1": {Name: "node1", Conn: &testMockSubConn{}},
		"node2": {Name: "node2", Conn: &testMockSubConn{}},
	}

	// 添加初始节点
	for _, node := range initialNodes {
		ch.Add(node)
	}

	// 新的节点集合
	newNodes := map[string]Node{
		"node2": {Name: "node2", Conn: &testMockSubConn{}},
		"node3": {Name: "node3", Conn: &testMockSubConn{}},
		"node4": {Name: "node4", Conn: &testMockSubConn{}},
	}

	// 执行全量替换
	ch.AllReplace(newNodes)

	// 验证节点替换结果
	key1 := "test_key_1"
	key2 := "test_key_2"
	key3 := "test_key_3"

	result1 := ch.Get(key1)
	result2 := ch.Get(key2)
	result3 := ch.Get(key3)

	// 验证获取的节点不为空且在新节点集合中
	assert.NotEmpty(t, result1)
	assert.NotEmpty(t, result2)
	assert.NotEmpty(t, result3)

	// 验证node1已被移除
	assert.NotEqual(t, "node1", result1.Name)
}

func TestConsistentHash_Consistency(t *testing.T) {
	ch := NewConsistentHash(3)

	// 添加初始节点
	node1 := Node{Name: "node1", Conn: &testMockSubConn{}}
	node2 := Node{Name: "node2", Conn: &testMockSubConn{}}
	node3 := Node{Name: "node3", Conn: &testMockSubConn{}}
	ch.Add(node1)
	ch.Add(node2)
	ch.Add(node3)

	// 生成30个测试key
	testKeys := make([]string, 30)
	for i := 0; i < 30; i++ {
		testKeys[i] = fmt.Sprintf("test_key_%d", i)
	}

	// 记录每个key初始分配的节点
	initialMappings := make(map[string]Node)
	for _, key := range testKeys {
		initialMappings[key] = ch.Get(key)
	}

	// 删除node1
	ch.Remove(node1)

	// 验证key的重新分配
	for _, key := range testKeys {
		originalNode := initialMappings[key]
		newNode := ch.Get(key)

		if originalNode.Name == "node1" {
			assert.NotEqual(t, "node1", newNode.Name)
			assert.NotEmpty(t, newNode)
		} else {
			assert.Equal(t, originalNode.Name, newNode.Name)
		}
	}
}
