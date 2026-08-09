package service

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type StateMachineService struct{}

// allowedTransitions 合法流转表
var allowedTransitions = map[string]map[string]bool{
	"draft":     {"reviewing": true},
	"reviewing": {"published": true, "draft": true},
	"published": {"archived": true},
	"archived":  {"draft": true},
}

// Transition 统一配置状态流转。table 为 pm_entity_types/pm_field_defs/pm_workflow_defs/pm_relation_types/pm_state_defs。
func (s *StateMachineService) Transition(db *gorm.DB, table string, id uint, from, to string) error {
	if !allowedTransitions[from][to] {
		return fmt.Errorf("非法流转: %s → %s", from, to)
	}
	var current string
	if err := db.Table(table).Select("status").Where("id = ?", id).Scan(&current).Error; err != nil {
		return err
	}
	if current != from {
		return fmt.Errorf("状态不一致: 期望 %s，实际 %s", from, current)
	}
	return db.Table(table).Where("id = ?", id).Update("status", to).Error
}

// DeleteDraft 仅 draft 状态可删除（通用表级删除）
func (s *StateMachineService) DeleteDraft(db *gorm.DB, table string, id uint) error {
	var current string
	if err := db.Table(table).Select("status").Where("id = ?", id).Scan(&current).Error; err != nil {
		return err
	}
	if current != "draft" {
		return errors.New("仅 draft 状态可删除，请先归档")
	}
	return db.Table(table).Where("id = ?", id).Delete(nil).Error
}
