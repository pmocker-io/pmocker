package mcpTool

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestRiskMatrix(t *testing.T) {
	mockUpstreamServer(t, "/api/pmocker/risk/matrix", map[string]any{
		"code": 0,
		"data": map[string]any{"id": 1},
		"msg":  "ok",
	})

	tool := &PmockerRiskMatrix{}
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"project_id": "1",
		"threshold":  "9",
	}

	result, err := tool.Handle(mockMCPContext(), req)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}
