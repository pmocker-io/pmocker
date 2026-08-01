package mcpTool

import (
	"testing"
)

func TestMockUpstreamServer(t *testing.T) {
	mockUpstreamServer(t, "/api/test", map[string]any{"code": 0, "data": map[string]any{"ok": true}, "msg": "ok"})

	resp, err := postUpstream[map[string]any](mockMCPContext(), "/api/test", map[string]any{})
	if err != nil {
		t.Fatalf("postUpstream failed: %v", err)
	}
	if resp.Data["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp.Data["ok"])
	}
}
