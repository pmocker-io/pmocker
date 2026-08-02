package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

type TaskLinkService struct{}

func (s *TaskLinkService) CreateTaskLink(link pmocker.PMTaskLink) error {
	if link.SrcTaskID == link.DstTaskID {
		return nil
	}
	return global.GVA_DB.Create(&link).Error
}

func (s *TaskLinkService) DeleteTaskLink(id uint) error {
	return global.GVA_DB.Delete(&pmocker.PMTaskLink{}, id).Error
}

func (s *TaskLinkService) ListTaskLinks(projectID uint) ([]pmocker.PMTaskLink, error) {
	var links []pmocker.PMTaskLink
	err := global.GVA_DB.
		Joins("JOIN pm_entities e ON e.id = pm_task_links.src_task_id").
		Where("e.project_id = ?", projectID).
		Find(&links).Error
	return links, err
}
