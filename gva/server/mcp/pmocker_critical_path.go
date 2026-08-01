package mcpTool

import (
	"context"
	"errors"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func init() {
	RegisterTool(&CriticalPathAnalyzer{})
}

// CriticalPathAnalyzer 项目关键路径(CPM)分析工具
type CriticalPathAnalyzer struct{}

func (t *CriticalPathAnalyzer) New() mcp.Tool {
	return mcp.NewTool("pmocker_critical_path",
		mcp.WithDescription(`分析项目关键路径(CPM, Critical Path Method)。

**功能说明：**
- 基于项目任务依赖关系计算关键路径
- 识别决定项目最短工期的关键任务链
- 支持以列表(list)或图(graph)视图返回结果

**适用场景：**
- 评估项目工期与任务调度风险
- 识别延期会影响整体交付的关键任务
- 排查进度瓶颈`),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("项目ID，必填"),
		),
		mcp.WithString("view",
			mcp.Description("结果视图：list(列表，默认) | graph(图)"),
			mcp.DefaultString("list"),
		),
	)
}

func (t *CriticalPathAnalyzer) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	projectID, ok := args["project_id"].(string)
	if !ok || strings.TrimSpace(projectID) == "" {
		return nil, errors.New("参数错误：project_id 必须是非空字符串")
	}

	view := "list"
	if value, ok := args["view"].(string); ok {
		if v := strings.TrimSpace(value); v != "" {
			view = v
		}
	}

	resp, err := postUpstream[any](ctx, "/api/pmocker/schedule/cpm", map[string]any{
		"project_id": projectID,
		"view":       view,
	})
	if err != nil {
		return nil, err
	}

	return textResultWithJSON("关键路径(CPM)分析结果：", resp.Data)
}
