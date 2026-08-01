package mcpTool

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestQueryIssues(t *testing.T) {
	mockUpstreamServer(t, "/api/pmocker/issue/list", map[string]any{
		"code": 0,
		"data": map[string]any{"id": 1},
		"msg":  "ok",
	})

	tool := &PmockerQueryIssues{}
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"project_id": "1",
		"status":     "open",
		"severity":   "critical",
		"assignee":   "alice",
	}

	result, err := tool.Handle(mockMCPContext(), req)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}
