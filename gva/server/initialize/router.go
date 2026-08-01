package initialize

import (
	"net/http"
	"os"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/docs"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/flipped-aurora/gin-vue-admin/server/router"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type justFilesFilesystem struct {
	fs http.FileSystem
}

func (fs justFilesFilesystem) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err == nil && stat.IsDir() {
		return nil, os.ErrPermission
	}

	return f, nil
}

// 初始化总路由

func Routers() *gin.Engine {
	Router := gin.New()
	// RequestMeta 必须最先：保证 panic 日志与 X-Request-Id 响应头都带 request_id
	Router.Use(middleware.RequestMeta())
	// 使用自定义的 Recovery 中间件，记录 panic 并入库
	Router.Use(middleware.GinRecovery(true))
	// 全局访问日志 + 唯一 body/resp 捕获点（供 OperationRecord 复用）
	Router.Use(middleware.AccessLog())
	if gin.Mode() == gin.DebugMode {
		Router.Use(gin.Logger())
	}

	systemRouter := router.RouterGroupApp.System
	exampleRouter := router.RouterGroupApp.Example
	mediaRouter := router.RouterGroupApp.Media
	// 如果 dist 目录存在，服务前端静态文件（pmocker 实例模式）
	if _, err := os.Stat("./dist"); err == nil {
		Router.StaticFile("/favicon.ico", "./dist/favicon.ico")
		Router.StaticFile("/logo.png", "./dist/logo.png")
		Router.Static("/assets", "./dist/assets")
		Router.StaticFile("/", "./dist/index.html")
		// SPA fallback：所有未匹配的非 API GET 请求返回 index.html，避免前端路由刷新时 404
		Router.NoRoute(func(c *gin.Context) {
			if c.Request.Method != http.MethodGet {
				return
			}
			if strings.HasPrefix(c.Request.URL.Path, global.GVA_CONFIG.System.RouterPrefix+"/") {
				return
			}
			c.File("./dist/index.html")
		})
	}

	Router.Use(middleware.UploadResponseHeaders(global.GVA_CONFIG.Local.StorePath))
	Router.StaticFS(global.GVA_CONFIG.Local.StorePath, justFilesFilesystem{http.Dir(global.GVA_CONFIG.Local.StorePath)})
	// Router.Use(middleware.LoadTls())  // 如果需要使用https 请打开此中间件 然后前往 core/server.go 将启动模式 更变为 Router.RunTLS("端口","你的cre/pem文件","你的key文件")
	// 跨域，如需跨域可以打开下面的注释
	// Router.Use(middleware.Cors()) // 直接放行全部跨域请求
	// Router.Use(middleware.CorsByRules()) // 按照配置的规则放行跨域请求
	docs.SwaggerInfo.BasePath = global.GVA_CONFIG.System.RouterPrefix
	Router.GET(global.GVA_CONFIG.System.RouterPrefix+"/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	logger.Bg().Mod("system").Info("register swagger handler")
	// 方便统一添加路由组前缀 多服务器上线使用

	PublicGroup := Router.Group(global.GVA_CONFIG.System.RouterPrefix)
	PrivateGroup := Router.Group(global.GVA_CONFIG.System.RouterPrefix)

	PrivateGroup.Use(middleware.JWTAuth()).Use(middleware.MustChangePwdGuard()).Use(middleware.CasbinHandler()).Use(middleware.DataScope())

	{
		// 健康监测
		PublicGroup.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, "ok")
		})
	}
	{
		systemRouter.InitBaseRouter(PublicGroup) // 注册基础功能路由 不做鉴权
		systemRouter.InitInitRouter(PublicGroup) // 自动初始化相关
	}

	{
		systemRouter.InitApiRouter(PrivateGroup, PublicGroup)               // 注册功能api路由
		systemRouter.InitJwtRouter(PrivateGroup)                            // jwt相关路由
		systemRouter.InitUserRouter(PrivateGroup)                           // 注册用户路由
		systemRouter.InitMenuRouter(PrivateGroup)                           // 注册menu路由
		systemRouter.InitSystemRouter(PrivateGroup)                         // system相关路由
		systemRouter.InitSysVersionRouter(PrivateGroup)                     // 发版相关路由
		systemRouter.InitCasbinRouter(PrivateGroup)                         // 权限相关路由
		systemRouter.InitAuthorityRouter(PrivateGroup)                      // 注册角色路由
		systemRouter.InitSysDepartmentRouter(PrivateGroup)                  // 注册部门路由
		systemRouter.InitSysPositionRouter(PrivateGroup)                    // 注册岗位路由
		systemRouter.InitSysDataAccessLogRouter(PrivateGroup)               // 数据权限审计日志
		systemRouter.InitSysDictionaryRouter(PrivateGroup)                  // 字典管理
		systemRouter.InitSysOperationRecordRouter(PrivateGroup)             // 操作记录
		systemRouter.InitSysDictionaryDetailRouter(PrivateGroup)            // 字典详情管理
		systemRouter.InitAuthorityBtnRouterRouter(PrivateGroup)             // 按钮权限管理
		systemRouter.InitSysExportTemplateRouter(PrivateGroup, PublicGroup) // 导出模板
		systemRouter.InitSysParamsRouter(PrivateGroup, PublicGroup)         // 参数管理
		systemRouter.InitSysErrorRouter(PrivateGroup, PublicGroup)          // 错误日志
		systemRouter.InitLoginLogRouter(PrivateGroup)                       // 登录日志
		systemRouter.InitSecurityConfigRouter(PrivateGroup)                 // 安全配置
		systemRouter.InitApiTokenRouter(PrivateGroup)                       // apiToken签发
		systemRouter.InitTimedTaskRouter(PrivateGroup)                      // 定时任务
		systemRouter.InitLogViewerRouter(PrivateGroup)                      // 文件日志查看
		exampleRouter.InitCustomerRouter(PrivateGroup)                      // 客户路由
		mediaRouter.InitFileUploadAndDownloadRouter(PrivateGroup)           // 文件上传下载功能路由
		mediaRouter.InitAttachmentCategoryRouterRouter(PrivateGroup)        // 媒体分类
		mediaRouter.InitMediaUploadRouter(PrivateGroup)                     // 大文件上传
	}

	//插件路由安装
	InstallPlugin(PrivateGroup, PublicGroup, Router)

	// 注册业务路由
	initBizRouter(PrivateGroup, PublicGroup)

	global.GVA_ROUTERS = Router.Routes()

	logger.Bg().Mod("system").Info("router register success")
	return Router
}
