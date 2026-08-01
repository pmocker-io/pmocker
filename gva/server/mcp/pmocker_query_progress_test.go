package mcpTool

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestQueryProgress(t *testing.T) {
	mockUpstreamServer(t, "/api/pmocker/schedule/listTasks", map[string]any{
		"code": 0,
		"data": map[string]any{
			"list":  []map[string]any{{"id": "1", "name": "任务A"}},
			"total": 1,
		},
		"msg": "ok",
	})
	tool := &ProgressQueryer{}
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"project_id": "1",
		"scope":      "active",
	}
	result, err := tool.Handle(mockMCPContext(), req)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}
