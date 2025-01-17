package testcase

import (
	"math/rand"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func Benchmark_ReadWriteHttp(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	// 复用 HTTP 客户端
	client := &http.Client{}

	b.ResetTimer()
	b.SetParallelism(1000)

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			resp, err := client.Get("http://localhost:8080/write?user=" + strconv.Itoa(random.Intn(1000)))
			if err != nil {
				b.Error(err)
				continue
			}
			resp.Body.Close() // 确保关闭响应体
			b.Logf("write resp status: %d", resp.StatusCode)
		}
	})

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			resp, err := client.Get("http://localhost:8080/read?user=" + strconv.Itoa(random.Intn(1000)))
			if err != nil {
				b.Error(err)
				continue
			}
			resp.Body.Close() // 确保关闭响应体
			b.Logf("read resp status: %d", resp.StatusCode)
		}
	})
}
