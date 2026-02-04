package providers

import (
	"fmt"
	"orderease/config"
	"orderease/database"
	ordercontextHandlers "orderease/contexts/ordercontext/application/handlers"
	"orderease/services"
	"orderease/tasks"
	"orderease/utils/log2"

	"gorm.io/gorm"
)

// Container 依赖注入容器
type Container struct {
	Config           *config.Config
	DB               *gorm.DB
	Logger           *log2.Logger
	Handler          *ordercontextHandlers.Handler
	TempTokenService *services.TempTokenService
	CleanupTask      *tasks.CleanupTask
}

// InitializeApp 初始化应用程序（手动依赖注入）
func InitializeApp(configPath string) (*Container, error) {
	container := &Container{}

	// 1. 加载配置
	if err := config.LoadConfig(configPath); err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}
	container.Config = &config.AppConfig

	// 2. 初始化日志
	log2.InitLogger()
	container.Logger = log2.GetLogger()

	// 3. 初始化数据库
	db, err := database.Init()
	if err != nil {
		return nil, fmt.Errorf("数据库初始化失败: %w", err)
	}
	container.DB = db

	// 4. 创建 Handler（Handler 的 NewHandler 会创建所有依赖）
	container.Handler = ordercontextHandlers.NewHandler(db)

	// 5. 创建 Application Services（需要从 Handler 获取或单独创建）
	tempTokenService := services.NewTempTokenService(db)
	container.TempTokenService = tempTokenService

	cleanupTask := tasks.NewCleanupTask(db)
	container.CleanupTask = cleanupTask

	return container, nil
}
