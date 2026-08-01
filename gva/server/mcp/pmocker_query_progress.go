package mcpTool

import (
	"context"
	"errors"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
)

func init() {
	RegisterTool(&ProgressQueryer{})
}

type ProgressQueryer struct{}

func (t *ProgressQueryer) New() mcp.Tool {
	return mcp.NewTool("pmocker_query_progress",
		mcp.WithDescription(`查询项目进度，返回任务列表用于进度分析。scope=active 返回进行中任务，scope=delayed 返回延迟任务。

**功能说明：**
- 调用 /api/pmocker/schedule/listTasks 获取任务列表
- scope=all：返回项目全部任务（默认）
- scope=active：仅返回进行中任务，用于跟进当前进度
- scope=delayed：仅返回延迟任务，用于风险预警

**适用场景：**
- 项目进度周报/日报
- 识别延迟任务并跟进
- 分析任务分布与瓶颈`),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("项目ID"),
		),
		mcp.WithString("scope",
			mcp.Description("查询范围: all(全部) | active(进行中) | delayed(延迟)"),
			mcp.DefaultString("all"),
		),
	)
}

func (t *ProgressQueryer) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	projectID, ok := args["project_id"].(string)
	if !ok || projectID == "" {
		return nil, errors.New("project_id 参数是必需的")
	}

	scope := "all"
	if value, ok := args["scope"].(string); ok && value != "" {
		scope = value
	}

	query := url.Values{}
	query.Set("project_id", projectID)
	query.Set("scope", scope)

	resp, err := getUpstream[map[string]any](ctx, "/api/pmocker/schedule/listTasks", query)
	if err != nil {
		return nil, err
	}

	return textResultWithJSON("项目进度查询结果：", resp.Data)
}
