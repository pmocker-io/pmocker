package initialize

import (
	"os"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	sysService "github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
)

// AutoInitIfEmpty 在 PMocker 实例模式下自动初始化数据库
// 当环境变量 PMOCKER_AUTO_INIT=1 且数据库中无管理员用户时，
// 自动调用 InitDB 完成种子数据初始化
func AutoInitIfEmpty() {
	if os.Getenv("PMOCKER_AUTO_INIT") != "1" {
		return
	}
	if global.GVA_DB == nil {
		return
	}

	// 检查是否已有管理员用户
	var count int64
	if err := global.GVA_DB.Model(&sysModel.SysUser{}).Count(&count).Error; err != nil {
		logger.Bg().Mod("auto-init").Err(err).Error("检查用户表失败")
		return
	}
	if count > 0 {
		logger.Bg().Mod("auto-init").Info("数据库已有用户数据，跳过自动初始化")
		return
	}

	// 从配置构造 InitDB 请求
	conf := request.InitDB{
		AdminPassword: getAdminPassword(),
		DBType:        global.GVA_CONFIG.System.DbType,
	}

	// SQLite 需要填充 DBPath 和 DBName
	if conf.DBType == "sqlite" {
		conf.DBPath = global.GVA_CONFIG.Sqlite.Path
		conf.DBName = global.GVA_CONFIG.Sqlite.Dbname
	}

	logger.Bg().Mod("auto-init").Info("PMocker 自动初始化数据库...")
	initDBService := &sysService.InitDBService{}
	if err := initDBService.InitDB(conf); err != nil {
		logger.Bg().Mod("auto-init").Err(err).Error("自动初始化数据库失败")
		return
	}
	// PMocker 实例模式禁用验证码（自动化测试/演示场景）
	if err := global.GVA_DB.Table("sys_security_config").Where("id = ?", 1).Update("captcha_open", 99999).Error; err != nil {
		logger.Bg().Mod("auto-init").Err(err).Error("禁用验证码失败")
	}
	logger.Bg().Mod("auto-init").Info("PMocker 自动初始化数据库完成")
}

// getAdminPassword 从环境变量获取管理员密码，默认 123456
func getAdminPassword() string {
	pwd := os.Getenv("PMOCKER_ADMIN_PASSWORD")
	if pwd == "" {
		pwd = "123456"
	}
	return pwd
}
