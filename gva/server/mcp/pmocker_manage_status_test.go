package mcpTool

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestManageStatus(t *testing.T) {
	mockUpstreamServer(t, "/api/pmocker/requirement/submitReview", map[string]any{
		"code": 0,
		"data": map[string]any{"new_status": "in_progress"},
		"msg":  "状态流转成功",
	})
	tool := &StatusManager{}
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"entity_type": "requirement",
		"entity_id":   "1",
		"action":      "submit",
		"comment":     "提交评审",
	}
	result, err := tool.Handle(mockMCPContext(), req)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}
