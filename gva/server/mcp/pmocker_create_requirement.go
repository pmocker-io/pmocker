package mcpTool

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func init() {
	RegisterTool(&PmockerCreateRequirement{})
}

type PmockerCreateRequirement struct{}

func (t *PmockerCreateRequirement) New() mcp.Tool {
	return mcp.NewTool("pmocker_create_requirement",
		mcp.WithDescription(`创建 PMocker 需求记录，用于在指定模块下登记一条新的需求条目。

**功能说明：**
- 向 GVA 主服务 POST /api/pmocker/requirement/create 提交需求
- priority 缺省为 medium，可传 high|medium|low
- description、module 为可选项，仅在非空时随请求体一并发送`),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("需求标题"),
		),
		mcp.WithString("description",
			mcp.Description("需求详细描述"),
		),
		mcp.WithString("priority",
			mcp.Description("优先级：high|medium|low，默认 medium"),
			mcp.DefaultString("medium"),
		),
		mcp.WithString("module",
			mcp.Description("所属模块名称"),
		),
	)
}

func (t *PmockerCreateRequirement) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	title, ok := args["title"].(string)
	if !ok || title == "" {
		return nil, fmt.Errorf("title 参数必填")
	}

	priority := "medium"
	if value, ok := args["priority"].(string); ok && value != "" {
		priority = value
	}

	body := map[string]any{
		"title":    title,
		"priority": priority,
	}
	if description, ok := args["description"].(string); ok && description != "" {
		body["description"] = description
	}
	if module, ok := args["module"].(string); ok && module != "" {
		body["module"] = module
	}

	resp, err := postUpstream[map[string]any](ctx, "/api/pmocker/requirement/create", body)
	if err != nil {
		return nil, fmt.Errorf("创建需求失败: %w", err)
	}

	return textResultWithJSON("需求创建结果：", resp.Data)
}
