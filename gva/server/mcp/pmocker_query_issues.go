package mcpTool

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
)

func init() {
	RegisterTool(&PmockerQueryIssues{})
}

type PmockerQueryIssues struct{}

func (t *PmockerQueryIssues) New() mcp.Tool {
	return mcp.NewTool("pmocker_query_issues",
		mcp.WithDescription(`查询 PMocker 项目下的 Issue 列表，支持按状态、严重程度、指派人过滤。

**功能说明：**
- 向 GVA 主服务 GET /api/pmocker/issue/list 拉取 Issue
- project_id 为必填，其余过滤项可选，仅在非空时作为 query 参数透传`),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("项目 ID"),
		),
		mcp.WithString("status",
			mcp.Description("按状态过滤，如 open|closed"),
		),
		mcp.WithString("severity",
			mcp.Description("按严重程度过滤，如 critical|major|minor"),
		),
		mcp.WithString("assignee",
			mcp.Description("按指派人过滤"),
		),
	)
}

func (t *PmockerQueryIssues) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	projectID, ok := args["project_id"].(string)
	if !ok || projectID == "" {
		return nil, fmt.Errorf("project_id 参数必填")
	}

	query := url.Values{}
	query.Set("project_id", projectID)
	if status, ok := args["status"].(string); ok && status != "" {
		query.Set("status", status)
	}
	if severity, ok := args["severity"].(string); ok && severity != "" {
		query.Set("severity", severity)
	}
	if assignee, ok := args["assignee"].(string); ok && assignee != "" {
		query.Set("assignee", assignee)
	}

	resp, err := getUpstream[map[string]any](ctx, "/api/pmocker/issue/list", query)
	if err != nil {
		return nil, fmt.Errorf("查询 Issue 失败: %w", err)
	}

	return textResultWithJSON("Issue 查询结果：", resp.Data)
}
