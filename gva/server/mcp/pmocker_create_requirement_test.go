package mcpTool

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestCreateRequirement(t *testing.T) {
	mockUpstreamServer(t, "/api/pmocker/requirement/create", map[string]any{
		"code": 0,
		"data": map[string]any{"id": 1},
		"msg":  "ok",
	})

	tool := &PmockerCreateRequirement{}
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"title":       "测试需求",
		"description": "描述内容",
		"priority":    "high",
		"module":      "用户中心",
	}

	result, err := tool.Handle(mockMCPContext(), req)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}
