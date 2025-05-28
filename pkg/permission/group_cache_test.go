package permission_test

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/golang/groupcache"
)

/*
 * ----------------------- 公共工具 -----------------------
 */

var once sync.Once

// 所有节点共用的 Group 定义
func initSquareGroup() {
	once.Do(func() {
		groupcache.NewGroup("square", 16<<20, groupcache.GetterFunc(
			func(_ context.Context, key string, dest groupcache.Sink) error {
				n, _ := strconv.ParseInt(key, 10, 64)
				return dest.SetString(strconv.FormatInt(n*n, 10))
			}))
	})
}

/* ---------------- 节点结构 ---------------- */

type node struct {
	addr  string
	pool  *groupcache.HTTPPool
	srv   *http.Server
	index int // 方便打日志
}

func (n *node) close() { _ = n.srv.Close() }

/* ---------------- 节点初始化 ---------------- */

func newNode(idx int, self string, peers []string) *node {
	// 1. 创建 HTTPPool 并注入 peers
	pool := groupcache.NewHTTPPool(self)
	pool.Set(peers...) // peers 包含自己

	// 2. 确保 Group 已注册（每进程只会执行一次）
	initSquareGroup()

	// 3. 启动 HTTP Server 处理 groupcache RPC
	listen := self[len("http://"):]
	srv := &http.Server{Addr: listen, Handler: pool}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("node-%d listen %s err=%v", idx, self, err)
		}
	}()

	return &node{addr: self, pool: pool, srv: srv, index: idx}
}

/* ---------------- 单元测试入口 ---------------- */

func TestNodeAsServerAndClient(t *testing.T) {
	peers := []string{
		"http://127.0.0.1:8101",
		"http://127.0.0.1:8102",
		"http://127.0.0.1:8103",
	}

	// 初始化 3 个节点
	nodes := []*node{
		newNode(1, peers[0], peers),
		newNode(2, peers[1], peers),
		newNode(3, peers[2], peers),
	}
	defer func() {
		for _, n := range nodes {
			n.close()
		}
	}()

	time.Sleep(100 * time.Millisecond) // 确保全部监听起来

	/* ------------------------------------------
	   每个节点各自并发执行一次业务查询：
	   1) 实际 owner 可能是自己，也可能是别的节点
	   2) 结果统一应该是 n*n
	------------------------------------------ */
	var wg sync.WaitGroup
	g := groupcache.GetGroup("square")

	keys := []string{"2", "5", "9"} // 三个不同 key，提高命中不同 owner 的概率
	for i, n := range nodes {
		wg.Add(1)
		go func(nd *node, key string) {
			defer wg.Done()
			var v string
			if err := g.Get(context.Background(), key, groupcache.StringSink(&v)); err != nil {
				t.Errorf("node-%d Get(%s) error: %v", nd.index, key, err)
				return
			}
			fmt.Printf("node-%d  square(%s) = %s\n", nd.index, key, v)
		}(n, keys[i])
	}
	wg.Wait()

	// 只要全部打印成功即通过；可通过 -v 观察输出
}

func TestMultiGroupIndependence(t *testing.T) {
	g1 := groupcache.NewGroup("user-name", 8<<20, groupcache.GetterFunc(
		func(_ context.Context, key string, dest groupcache.Sink) error {
			return dest.SetString("user:" + key)
		}))

	g2 := groupcache.NewGroup("user-age", 8<<20, groupcache.GetterFunc(
		func(_ context.Context, key string, dest groupcache.Sink) error {
			return dest.SetString(strconv.FormatInt(int64(len(key)), 10))
		}))

	var name string
	if err := g1.Get(nil, "alice", groupcache.StringSink(&name)); err != nil {
		t.Fatal(err)
	}
	if name != "user:alice" {
		t.Fatal("unexpected")
	}

	var age []byte
	if err := g2.Get(nil, "alice", groupcache.AllocatingByteSliceSink(&age)); err != nil {
		t.Fatal(err)
	}
	if string(age) != "5" {
		t.Fatal("unexpected")
	}
}

func TestModifyReturnedBytes(t *testing.T) {
	g := groupcache.NewGroup("blob", 32<<20, groupcache.GetterFunc(
		func(_ context.Context, key string, dest groupcache.Sink) error {
			return dest.SetBytes(bytes.Repeat([]byte{1}, 10))
		}))

	var b []byte
	if err := g.Get(nil, "a", groupcache.AllocatingByteSliceSink(&b)); err != nil {
		t.Fatal(err)
	}
	b[0] = 42 // 修改副本 OK

	var again []byte
	_ = g.Get(nil, "a", groupcache.AllocatingByteSliceSink(&again))
	if again[0] == 42 {
		t.Fatal("should not be mutated")
	}
}

func TestStats(t *testing.T) {
	g := groupcache.NewGroup("stats", 1<<20, groupcache.GetterFunc(
		func(_ context.Context, key string, dest groupcache.Sink) error {
			return dest.SetString("x")
		}))
	for i := 0; i < 100; i++ {
		var s string
		_ = g.Get(context.Background(), "k", groupcache.StringSink(&s))
	}
	st := g.Stats
	if int64(st.CacheHits) == int64(99) {
		t.Log("stats OK")
	} else {
		t.Fatalf("%+v", st)
	}
}
