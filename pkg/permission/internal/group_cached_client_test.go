//go:build manual

package internal

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang/groupcache"
)

/*****************************************************************
 *                      子进程代码                               *
 *  用 GO_WANT_HELPER_PROCESS 环境变量把同一 test 二进制当作“节点”
 *****************************************************************/
func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return // 不是 helper 模式，直接返回让 go test 正常流程继续
	}

	addr := os.Getenv("ADDR") // 本节点地址，如 http://127.0.0.1:8301
	peers := strings.Split(os.Getenv("PEERS"), ",")
	key := os.Getenv("KEY") // 本节点周期查询的 key

	/* ---------- 1. 初始化 HTTPPool & Group ---------- */
	pool := groupcache.NewHTTPPool(addr)
	pool.Set(peers...)

	groupcache.NewGroup("square", 16<<20, groupcache.GetterFunc(
		func(_ context.Context, k string, dest groupcache.Sink) error {
			n, _ := strconv.ParseInt(k, 10, 64)
			return dest.SetString(strconv.FormatInt(n*n, 10))
		}))

	/* ---------- 2. 启动 HTTPServer ------------------ */
	go func() {
		if err := http.ListenAndServe(addr[len("http://"):], pool); err != nil {
			log.Fatalf("listen %s: %v", addr, err)
		}
	}()

	// 向父进程发送“READY”信号
	fmt.Println("READY")
	os.Stdout.Sync()

	/* ---------- 3. 业务循环，证明自己也是客户端 -------- */
	g := groupcache.GetGroup("square")
	for {
		var v string
		_ = g.Get(context.Background(), key, groupcache.StringSink(&v))
		time.Sleep(2 * time.Second)
	}
}

/*****************************************************************
 *                         父进程                                *
 *****************************************************************/
func TestClusterWithProcesses(t *testing.T) {
	peerAddrs := []string{
		"http://127.0.0.1:8301",
		"http://127.0.0.1:8302",
		"http://127.0.0.1:8303",
	}
	keys := []string{"2", "5", "9"} // 三个节点各自查询的 key，随意

	readyCh := make(chan struct{}, len(peerAddrs))
	var procs []*exec.Cmd

	/* ---------- 1. 启动 3 个子进程 ------------------- */
	for i, addr := range peerAddrs {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess")
		cmd.Env = append(os.Environ(),
			"GO_WANT_HELPER_PROCESS=1",
			"ADDR="+addr,
			"PEERS="+strings.Join(peerAddrs, ","),
			"KEY="+keys[i],
		)

		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			t.Fatalf("start %s: %v", addr, err)
		}
		procs = append(procs, cmd)

		// 读取子进程 stdout，遇到 READY 就发信号
		go func(a string, r io.Reader) {
			sc := bufio.NewScanner(r)
			for sc.Scan() {
				if strings.Contains(sc.Text(), "READY") {
					fmt.Printf("[%s] ready\n", a)
					readyCh <- struct{}{}
					// 只要收到一次 READY 就够了，后续输出忽略
				}
			}
		}(addr, stdout)

		// stderr 直接透传，方便调试
		go io.Copy(os.Stderr, stderr)
	}

	/* ---------- 2. 等待全部节点就绪 ------------------ */
	for i := 0; i < len(peerAddrs); i++ {
		<-readyCh
	}
	time.Sleep(200 * time.Millisecond) // 给 listener 最后一点缓冲时间

	/* ---------- 3. 父进程自身做一次 Get -------------- */
	clientPool := groupcache.NewHTTPPool("http://tester")
	clientPool.Set(peerAddrs...)

	// 注册一个同名 Group（子进程注册的是各自的，不会冲突）
	g := groupcache.GetGroup("square")
	if g == nil {
		g = groupcache.NewGroup("square", 16<<20, groupcache.GetterFunc(
			func(_ context.Context, k string, dest groupcache.Sink) error {
				return dest.SetString("unused")
			}))
	}

	var res string
	if err := g.Get(context.Background(), "12", groupcache.StringSink(&res)); err != nil {
		t.Fatalf("parent Get error: %v", err)
	}
	if res != "144" {
		t.Fatalf("expect 144, got %s", res)
	}
	t.Logf("parent square(12)=%s", res)

	/* ---------- 4. 结束子进程，收尾 ------------------ */
	for _, p := range procs {
		_ = p.Process.Kill()
		_ = p.Wait()
	}
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
