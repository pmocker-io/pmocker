package mcpTool

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
)

func init() {
	RegisterTool(&PmockerRiskMatrix{})
}

type PmockerRiskMatrix struct{}

func (t *PmockerRiskMatrix) New() mcp.Tool {
	return mcp.NewTool("pmocker_risk_matrix",
		mcp.WithDescription(`获取 PMocker 项目的风险矩阵，仅返回风险分值不低于阈值的项目。

**功能说明：**
- 向 GVA 主服务 GET /api/pmocker/risk/matrix 拉取风险矩阵
- threshold 缺省为 9，用于过滤低风险条目`),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("项目 ID"),
		),
		mcp.WithString("threshold",
			mcp.Description("风险阈值，仅返回分值不低于该阈值的风险，默认 9"),
			mcp.DefaultString("9"),
		),
	)
}

func (t *PmockerRiskMatrix) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	projectID, ok := args["project_id"].(string)
	if !ok || projectID == "" {
		return nil, fmt.Errorf("project_id 参数必填")
	}

	threshold := "9"
	if value, ok := args["threshold"].(string); ok && value != "" {
		threshold = value
	}

	query := url.Values{}
	query.Set("project_id", projectID)
	query.Set("threshold", threshold)

	resp, err := getUpstream[map[string]any](ctx, "/api/pmocker/risk/matrix", query)
	if err != nil {
		return nil, fmt.Errorf("获取风险矩阵失败: %w", err)
	}

	return textResultWithJSON("风险矩阵结果：", resp.Data)
}
