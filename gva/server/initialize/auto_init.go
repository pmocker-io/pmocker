package initialize

import (
	"context"
	"errors"
	"fmt"
	"os"

	adapter "github.com/casbin/gorm-adapter/v3"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	autoModel "github.com/flipped-aurora/gin-vue-admin/server/plugin/ai/model"
	sysService "github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
	"gorm.io/gorm"
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

// AutoInitPostPlugins 在插件初始化（Routers）之后执行，注册 MCP 动态工具、Casbin 规则和签发 API Token。
// 这些操作依赖 PMocker 插件已注册的 /pmocker/* API，必须在 Routers() 之后调用。
func AutoInitPostPlugins() {
	if os.Getenv("PMOCKER_AUTO_INIT") != "1" {
		return
	}
	if global.GVA_DB == nil {
		return
	}

	// 加载 PMocker 插件的 MCP 动态工具定义（创建 MCP 记录 + 绑定 PMocker API）
	if err := loadPMockerDynamicTools(); err != nil {
		logger.Bg().Mod("auto-init").Err(err).Error("加载 MCP 动态工具失败")
	}

	// 为 PMocker API 添加 Casbin 权限规则（authorityId=888），否则 API Token 调用返回"权限不足"
	if err := insertPMockerCasbinRules(); err != nil {
		logger.Bg().Mod("auto-init").Err(err).Error("插入 PMocker Casbin 规则失败")
	}

	// 将 PMocker 菜单授予管理员角色（authorityId=888），否则实例模式下菜单默认不可见，需手动在后台勾选
	if err := grantPMockerMenusToAdmin(); err != nil {
		logger.Bg().Mod("auto-init").Err(err).Error("授予 PMocker 菜单权限失败")
	}

	// 签发长期 API Token 供 MCP 使用
	token, err := issueLongLivedToken()
	if err != nil {
		logger.Bg().Mod("auto-init").Err(err).Error("签发 MCP Token 失败")
	} else {
		os.Setenv("PMOCKER_MCP_TOKEN", token)
		// 写入文件供 pmocker inspect 读取（gva-server 工作目录即实例数据卷）
		if werr := os.WriteFile("mcp_token.txt", []byte(token), 0644); werr != nil {
			logger.Bg().Mod("auto-init").Err(werr).Error("写入 MCP Token 文件失败")
		}
		logger.Bg().Mod("auto-init").Info("MCP Token 已签发")
	}
}

// loadPMockerDynamicTools 创建 PMocker MCP 定义并绑定所有 PMocker API 为动态工具。
// MCP 进程启动时通过 /mcpApi/listBindingsPublic 接口自动加载这些绑定。
func loadPMockerDynamicTools() error {
	db := global.GVA_DB

	// AutoInitIfEmpty 早于 AI 插件 Gorm() 初始化，需先确保 MCP 表已创建
	if err := db.AutoMigrate(&autoModel.SysMcp{}, &autoModel.SysMcpApi{}); err != nil {
		return fmt.Errorf("auto migrate mcp tables: %w", err)
	}

	// 1. 创建或获取名为 "pmocker" 的 MCP 记录
	var mcp autoModel.SysMcp
	err := db.Where("name = ?", "pmocker").First(&mcp).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		mcp = autoModel.SysMcp{
			Name:        "pmocker",
			DisplayName: "PMocker 项目管理",
			Description: "PMocker 项目管理系统的 MCP 动态工具集，包含 9 大模块的 CRUD 操作",
			Status:      "enabled",
			Version:     "v1",
		}
		if err := db.Create(&mcp).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// 2. 查询所有 PMocker API（path 以 /pmocker/ 开头）
	var apis []sysModel.SysApi
	if err := db.Where("path LIKE ?", "/pmocker/%").Find(&apis).Error; err != nil {
		return err
	}
	if len(apis) == 0 {
		logger.Bg().Mod("auto-init").Info("未找到 PMocker API，跳过 MCP 动态工具注册")
		return nil
	}

	// 3. 为每个 API 创建绑定（跳过已存在的）
	bound := 0
	for _, api := range apis {
		var existing autoModel.SysMcpApi
		err := db.Where("mcp_id = ? AND api_id = ?", mcp.ID, api.ID).First(&existing).Error
		if err == nil {
			continue // 已存在绑定，跳过
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		binding := autoModel.SysMcpApi{
			McpID:   mcp.ID,
			ApiID:   api.ID,
			Enabled: true,
		}
		if err := db.Create(&binding).Error; err != nil {
			logger.Bg().Mod("auto-init").Err(err).Error("创建 MCP 绑定失败: " + api.Path)
			continue
		}
		bound++
	}
	logger.Bg().Mod("auto-init").Info("MCP 动态工具注册完成: " + intToStr(len(apis)) + " API, " + intToStr(bound) + " 新绑定")
	return nil
}

// issueLongLivedToken 为管理员用户签发有效期 -1（100年）的 API Token，供 MCP 进程调用后端 API。
func issueLongLivedToken() (string, error) {
	ctx := context.Background()
	db := global.GVA_DB

	// 查询管理员用户
	var admin sysModel.SysUser
	if err := db.Where("username = ?", "admin").First(&admin).Error; err != nil {
		return "", err
	}

	// 检查是否已有 PMocker MCP Token（避免重复签发）
	var existingToken sysModel.SysApiToken
	err := db.Where("user_id = ? AND remark = ?", admin.ID, "PMocker MCP Token").First(&existingToken).Error
	if err == nil && existingToken.Token != "" {
		return existingToken.Token, nil
	}

	// 签发新 Token
	apiToken := sysModel.SysApiToken{
		UserID:      admin.ID,
		AuthorityID: admin.AuthorityId,
		Remark:      "PMocker MCP Token",
	}
	tokenService := &sysService.ApiTokenService{}
	token, err := tokenService.CreateApiToken(ctx, apiToken, -1) // -1 = 100年有效期
	if err != nil {
		return "", err
	}
	return token, nil
}

// intToStr 避免 fmt.Sprintf 在初始化路径引入额外依赖
func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// insertPMockerCasbinRules 为所有 PMocker API 创建 Casbin 权限规则（authorityId=888）。
// 参考 log_viewer_seed.go 的模式：FirstOrCreate 确保规则幂等，FreshCasbin 使规则立即生效。
func insertPMockerCasbinRules() error {
	db := global.GVA_DB
	var apis []sysModel.SysApi
	if err := db.Where("path LIKE ?", "/pmocker/%").Find(&apis).Error; err != nil {
		return err
	}
	for _, api := range apis {
		rule := adapter.CasbinRule{Ptype: "p", V0: "888", V1: api.Path, V2: api.Method}
		if err := db.Where(rule).FirstOrCreate(&rule).Error; err != nil {
			return err
		}
	}
	// 刷新 Casbin 使规则立即生效
	casbinService := &sysService.CasbinService{}
	return casbinService.FreshCasbin()
}

// grantPMockerMenusToAdmin 将所有 PMocker 菜单授予管理员角色（authorityId=888）。
// 解决实例模式下每次重新初始化后菜单默认不可见、需手动在后台「角色权限」勾选的问题。
// 参考 log_viewer_seed.go 的 FirstOrCreate 幂等模式。
func grantPMockerMenusToAdmin() error {
	db := global.GVA_DB
	var menus []sysModel.SysBaseMenu
	if err := db.Where("name LIKE ?", "pmocker%").Find(&menus).Error; err != nil {
		return err
	}
	if len(menus) == 0 {
		logger.Bg().Mod("auto-init").Info("未找到 PMocker 菜单，跳过菜单授权")
		return nil
	}
	granted := 0
	for _, menu := range menus {
		menuRole := sysModel.SysAuthorityMenu{
			MenuId:      fmt.Sprint(menu.ID),
			AuthorityId: "888",
		}
		// FirstOrCreate 幂等：已存在的关联不重复插入
		result := db.Where(menuRole).FirstOrCreate(&menuRole)
		if result.Error != nil {
			logger.Bg().Mod("auto-init").Err(result.Error).Error("授予菜单权限失败: " + menu.Name)
			continue
		}
		if result.RowsAffected > 0 {
			granted++
		}
	}
	logger.Bg().Mod("auto-init").Info("PMocker 菜单授权完成: " + intToStr(granted) + " 新增, " + intToStr(len(menus)) + " 总计")
	return nil
}

// getAdminPassword 从环境变量获取管理员密码，默认 123456
func getAdminPassword() string {
	pwd := os.Getenv("PMOCKER_ADMIN_PASSWORD")
	if pwd == "" {
		pwd = "123456"
	}
	return pwd
}
