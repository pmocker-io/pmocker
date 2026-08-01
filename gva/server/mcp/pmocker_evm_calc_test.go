package mcpTool

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestEvmCalculator(t *testing.T) {
	mockUpstreamServer(t, "/api/pmocker/cost/evm", map[string]any{
		"code": 0,
		"data": map[string]any{"result": "test"},
		"msg":  "ok",
	})

	tool := &EvmCalculator{}
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"project_id": "1",
	}

	result, err := tool.Handle(mockMCPContext(), req)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}
