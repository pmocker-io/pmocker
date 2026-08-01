package mcpTool

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/spf13/viper"
)

// mockUpstreamServer 启动一个 mock HTTP 服务模拟 GVA 后端，返回预设 JSON 响应。
// 自动设置 global.GVA_CONFIG.MCP.UpstreamBaseURL 指向 mock 服务，供 postUpstream/getUpstream 调用。
func mockUpstreamServer(t *testing.T, path string, response any) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	t.Cleanup(server.Close)
	// 设置全局配置指向 mock 服务
	if global.GVA_VP == nil {
		global.GVA_VP = viper.New()
	}
	global.GVA_CONFIG.MCP.UpstreamBaseURL = server.URL
	return server
}

// mockMCPContext 返回带测试 token 的 context，用于 postUpstream/getUpstream 鉴权。
// context key 必须与 context.go 中的 authTokenContextKey 一致。
func mockMCPContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, authTokenContextKey, "test-token")
	return ctx
}
