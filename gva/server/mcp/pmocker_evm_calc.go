package mcpTool

import (
	"context"
	"errors"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func init() {
	RegisterTool(&EvmCalculator{})
}

// EvmCalculator 挣值管理(EVM, Earned Value Management)计算工具
type EvmCalculator struct{}

func (t *EvmCalculator) New() mcp.Tool {
	return mcp.NewTool("pmocker_evm_calc",
		mcp.WithDescription(`计算项目挣值管理(EVM)指标。

**功能说明：**
- 基于 PV/AC/EV 三大基准计算 SV、CV、SPI、CPI 等挣值指标
- 评估项目进度与成本绩效，预测完工偏差与估算
- 支持按指定日期截取快照进行核算

**适用场景：**
- 项目绩效分析与成本控制
- 进度偏差/成本偏差诊断
- 完工估算(EAC/ETC)预测`),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("项目ID，必填"),
		),
		mcp.WithString("as_of_date",
			mcp.Description("核算截止日期，可选，格式 YYYY-MM-DD；缺省取当前日期"),
		),
	)
}

func (t *EvmCalculator) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	projectID, ok := args["project_id"].(string)
	if !ok || strings.TrimSpace(projectID) == "" {
		return nil, errors.New("参数错误：project_id 必须是非空字符串")
	}

	body := map[string]any{
		"project_id": projectID,
	}
	if value, ok := args["as_of_date"].(string); ok {
		if v := strings.TrimSpace(value); v != "" {
			body["as_of_date"] = v
		}
	}

	resp, err := postUpstream[any](ctx, "/api/pmocker/cost/evm", body)
	if err != nil {
		return nil, err
	}

	return textResultWithJSON("挣值管理(EVM)计算结果：", resp.Data)
}
