package mcpTool

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestImpactAnalysis(t *testing.T) {
	mockUpstreamServer(t, "/api/pmocker/change/impactReport", map[string]any{
		"code": 0,
		"data": map[string]any{"result": "test"},
		"msg":  "ok",
	})

	tool := &ImpactAnalyzer{}
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"change_id": "1",
	}

	result, err := tool.Handle(mockMCPContext(), req)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}
