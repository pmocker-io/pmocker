package mcpTool

import (
	"context"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func init() {
	RegisterTool(&StatusManager{})
}

// actionToEndpoint maps (entity_type, action) to the backend API path
var actionToEndpoint = map[string]map[string]string{
	"requirement": {
		"submit":  "/api/pmocker/requirement/submitReview",
		"approve": "/api/pmocker/requirement/approve",
		"reject":  "/api/pmocker/requirement/reject",
	},
	"change": {
		"ccbReview": "/api/pmocker/change/ccbReview",
		"approve":   "/api/pmocker/change/approve",
		"reject":    "/api/pmocker/change/reject",
		"implement": "/api/pmocker/change/implement",
		"verify":    "/api/pmocker/change/verify",
		"close":     "/api/pmocker/change/close",
	},
	"issue": {
		"assign":  "/api/pmocker/issue/assign",
		"resolve": "/api/pmocker/issue/resolve",
		"close":   "/api/pmocker/issue/close",
		"reopen":  "/api/pmocker/issue/reopen",
	},
	"risk": {
		"assess": "/api/pmocker/risk/assess",
	},
}

type StatusManager struct{}

func (t *StatusManager) New() mcp.Tool {
	return mcp.NewTool("pmocker_manage_status",
		mcp.WithDescription(`触发工作流状态流转，根据实体类型自动路由到对应模块的状态流转 API。

**功能说明：**
- 根据 entity_type + action 自动路由到后端对应的状态流转接口
- 支持的实体类型：requirement(需求) | change(变更) | issue(问题) | risk(风险)
- 支持的动作随实体类型不同而不同（详见各 action 列表）

**路由规则：**
- requirement: submit(提交评审) / approve(通过) / reject(驳回)
- change: ccbReview(CCB评审) / approve(通过) / reject(驳回) / implement(实施) / verify(验证) / close(关闭)
- issue: assign(指派) / resolve(解决) / close(关闭) / reopen(重开)
- risk: assess(评估)

**适用场景：**
- 工作流自动化流转
- 批量状态推进
- 审批/驳回处理`),
		mcp.WithString("entity_type",
			mcp.Required(),
			mcp.Description("实体类型: requirement | change | issue | risk"),
		),
		mcp.WithString("entity_id",
			mcp.Required(),
			mcp.Description("实体ID"),
		),
		mcp.WithString("action",
			mcp.Required(),
			mcp.Description("操作动作: submit | approve | reject | close | resolve | implement | verify | assess | assign | reopen | ccbReview"),
		),
		mcp.WithString("comment",
			mcp.Description("备注/审批意见(可选)"),
		),
	)
}

func (t *StatusManager) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	entityType, ok := args["entity_type"].(string)
	if !ok || entityType == "" {
		return nil, errors.New("entity_type 参数是必需的")
	}
	entityID, ok := args["entity_id"].(string)
	if !ok || entityID == "" {
		return nil, errors.New("entity_id 参数是必需的")
	}
	action, ok := args["action"].(string)
	if !ok || action == "" {
		return nil, errors.New("action 参数是必需的")
	}

	actionMap, ok := actionToEndpoint[entityType]
	if !ok {
		return nil, fmt.Errorf("不支持的操作: %s/%s", entityType, action)
	}
	endpoint, ok := actionMap[action]
	if !ok {
		return nil, fmt.Errorf("不支持的操作: %s/%s", entityType, action)
	}

	body := map[string]any{"id": entityID}
	if comment, ok := args["comment"].(string); ok && comment != "" {
		body["comment"] = comment
	}

	resp, err := postUpstream[map[string]any](ctx, endpoint, body)
	if err != nil {
		return nil, err
	}

	return textResultWithJSON("状态流转结果：", resp.Data)
}
