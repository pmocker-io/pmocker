package mcpTool

import (
	"context"
	"net/url"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func init() {
	RegisterTool(&ImpactAnalyzer{})
}

// ImpactAnalyzer 变更影响分析工具，分析变更对需求、任务和成本的影响
type ImpactAnalyzer struct{}

func (t *ImpactAnalyzer) New() mcp.Tool {
	return mcp.NewTool("pmocker_impact_analysis",
		mcp.WithDescription(`分析变更对需求、任务和成本的影响。

**功能说明：**
- 基于变更项评估对关联需求、任务及成本的连锁影响
- 支持按 change_id 精确查询特定变更的影响报告
- 未指定 change_id 时返回整体变更影响汇总

**适用场景：**
- 变更评审前的风险预判
- 评估变更对项目范围、进度与成本的连带影响
- 变更决策辅助`),
		mcp.WithString("change_id",
			mcp.Description("变更ID，可选；提供时按指定变更查询影响报告"),
		),
	)
}

func (t *ImpactAnalyzer) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	query := url.Values{}
	if value, ok := args["change_id"].(string); ok {
		if v := strings.TrimSpace(value); v != "" {
			query.Set("change_id", v)
		}
	}

	resp, err := getUpstream[any](ctx, "/api/pmocker/change/impactReport", query)
	if err != nil {
		return nil, err
	}

	return textResultWithJSON("变更影响分析结果：", resp.Data)
}
