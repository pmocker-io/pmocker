package pmocker

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"github.com/pmocker-io/pmocker/pkg/pmocker/rbac"
)

// RBACService 三维权限判定器
type RBACService struct{}

// CheckPermission 三维权限判定
// ① gva 角色含 action 权限标识（v1 由 gva casbin 中间件负责）
// ② EPS 节点上 user 有对应角色（M3 阶段补充）
// ③ 实体当前状态允许该 action
func (s *RBACService) CheckPermission(ctx context.Context, userID uint, action rbac.Action, entityID uint) rbac.Decision {
	var entity pmocker.PMEntity
	if err := global.GVA_DB.WithContext(ctx).First(&entity, entityID).Error; err != nil {
		return rbac.Decision{Allowed: false, Reason: "entity not found"}
	}
	allowed, reason := s.checkEntityState(entity.Status, action, entity.OwnerID, userID)
	if !allowed {
		return rbac.Decision{Allowed: false, Reason: reason}
	}
	return rbac.Decision{Allowed: true}
}

// checkEntityState 实体状态权限检查
func (s *RBACService) checkEntityState(status string, action rbac.Action, ownerID *uint, userID uint) (bool, string) {
	switch status {
	case "draft":
		if action == rbac.ActionUpdate || action == rbac.ActionDelete {
			if ownerID == nil || *ownerID != userID {
				return false, "draft state: only owner can modify"
			}
		}
	case "reviewing":
		if action == rbac.ActionUpdate {
			return false, "reviewing state: only approvers can annotate"
		}
	case "published", "released":
		if action == rbac.ActionUpdate || action == rbac.ActionDelete {
			return false, "published state: read-only, changes via change management"
		}
	case "archived":
		if action == rbac.ActionUpdate {
			return false, "archived state: only PMO can restore"
		}
	}
	return true, ""
}
