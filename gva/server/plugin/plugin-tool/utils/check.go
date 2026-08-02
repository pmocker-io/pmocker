package utils

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
)

var (
	ApiMap  = make(map[string][]system.SysApi)
	MenuMap = make(map[string][]system.SysBaseMenu)
	DictMap = make(map[string][]system.SysDictionary)
	rw      sync.Mutex
)

func getPluginName() string {
	_, file, _, ok := runtime.Caller(2)
	pluginName := ""
	if ok {
		file = filepath.ToSlash(file)
		const key = "server/plugin/"
		if idx := strings.Index(file, key); idx != -1 {
			remain := file[idx+len(key):]
			parts := strings.Split(remain, "/")
			if len(parts) > 0 {
				pluginName = parts[0]
			}
		}
	}
	return pluginName
}

func RegisterApis(apis ...system.SysApi) {
	name := getPluginName()
	if name != "" {
		rw.Lock()
		ApiMap[name] = apis
		rw.Unlock()
	}

	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		for _, api := range apis {
			err := tx.Model(system.SysApi{}).Where("path = ? AND method = ? AND api_group = ? ", api.Path, api.Method, api.ApiGroup).FirstOrCreate(&api).Error
			if err != nil {
				zap.L().Error("注册API失败", zap.Error(err), zap.String("api", api.Path), zap.String("method", api.Method), zap.String("apiGroup", api.ApiGroup))
				return err
			}
		}
		return nil
	})
	if err != nil {
		zap.L().Error("注册API失败", zap.Error(err))
	}
}

func RegisterMenus(menus ...system.SysBaseMenu) {
	name := getPluginName()
	if name != "" {
		rw.Lock()
		MenuMap[name] = menus
		rw.Unlock()
	}

	parentMenu := menus[0]
	otherMenus := menus[1:]
	var err error
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		// 先查找父菜单是否存在
		var existingMenu system.SysBaseMenu
		result := tx.Where("name = ? ", parentMenu.Name).First(&existingMenu)
		if result.Error == nil {
			// 父菜单已存在，更新字段（保留已有 ID）
			parentMenu.ID = existingMenu.ID
			if parentMenu.Path != "" {
				existingMenu.Path = parentMenu.Path
			}
			if parentMenu.Component != "" {
				existingMenu.Component = parentMenu.Component
			}
			if parentMenu.Sort != 0 {
				existingMenu.Sort = parentMenu.Sort
			}
			if parentMenu.Meta.Title != "" {
				existingMenu.Meta.Title = parentMenu.Meta.Title
			}
			if parentMenu.Meta.Icon != "" {
				existingMenu.Meta.Icon = parentMenu.Meta.Icon
			}
			if parentMenu.Hidden != existingMenu.Hidden {
				existingMenu.Hidden = parentMenu.Hidden
			}
			err = tx.Save(&existingMenu).Error
		} else {
			// 父菜单不存在，创建新记录
			err = tx.Create(&parentMenu).Error
		}
		if err != nil {
			zap.L().Error("注册菜单失败", zap.Error(err))
			return errors.Wrap(err, "注册菜单失败")
		}
		pid := parentMenu.ID
		if result.Error == nil {
			pid = existingMenu.ID
		}
		for i := range otherMenus {
			otherMenus[i].ParentId = pid
			// 先查找子菜单是否存在
			var existingChild system.SysBaseMenu
			childResult := tx.Where("name = ? ", otherMenus[i].Name).First(&existingChild)
			if childResult.Error == nil {
				// 子菜单已存在，更新字段
				existingChild.Path = otherMenus[i].Path
				existingChild.Component = otherMenus[i].Component
				if otherMenus[i].Sort != 0 {
					existingChild.Sort = otherMenus[i].Sort
				}
				existingChild.ParentId = pid
				existingChild.Meta = otherMenus[i].Meta
				existingChild.Hidden = otherMenus[i].Hidden
				err = tx.Save(&existingChild).Error
			} else {
				// 子菜单不存在，创建新记录
				err = tx.Create(&otherMenus[i]).Error
			}
			if err != nil {
				zap.L().Error("注册菜单失败", zap.Error(err))
				return errors.Wrap(err, "注册菜单失败")
			}
		}
		return nil
	})
	if err != nil {
		zap.L().Error("注册菜单失败", zap.Error(err))
	}

}

func RegisterDictionaries(dictionaries ...system.SysDictionary) {
	name := getPluginName()
	if name != "" {
		rw.Lock()
		DictMap[name] = dictionaries
		rw.Unlock()
	}

	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		for _, dict := range dictionaries {
			details := dict.SysDictionaryDetails
			dict.SysDictionaryDetails = nil
			err := tx.Model(system.SysDictionary{}).Where("type = ?", dict.Type).FirstOrCreate(&dict).Error
			if err != nil {
				zap.L().Error("注册字典失败", zap.Error(err), zap.String("type", dict.Type))
				return err
			}
			for _, detail := range details {
				detail.SysDictionaryID = int(dict.ID)
				err = tx.Model(system.SysDictionaryDetail{}).Where("sys_dictionary_id = ? AND value = ?", dict.ID, detail.Value).FirstOrCreate(&detail).Error
				if err != nil {
					zap.L().Error("注册字典详情失败", zap.Error(err), zap.String("value", detail.Value))
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		zap.L().Error("注册字典失败", zap.Error(err))
	}
}

func Pointer[T any](in T) *T {
	return &in
}

func GetPluginData(pluginName string) ([]system.SysApi, []system.SysBaseMenu, []system.SysDictionary) {
	rw.Lock()
	defer rw.Unlock()
	return ApiMap[pluginName], MenuMap[pluginName], DictMap[pluginName]
}

