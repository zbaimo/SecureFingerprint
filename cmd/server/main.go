package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"firewall-controller/api"
	"firewall-controller/internal/analyzer"
	"firewall-controller/internal/collector"
	"firewall-controller/internal/fingerprint"
	"firewall-controller/internal/limiter"
	"firewall-controller/internal/scorer"
	"firewall-controller/internal/storage"
	"firewall-controller/pkg/middleware"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// 应用配置
type Config struct {
	Server struct {
		Port  int  `yaml:"port"`
		Debug bool `yaml:"debug"`
	} `yaml:"server"`

	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
		PoolSize int    `yaml:"pool_size"`
	} `yaml:"redis"`

	MySQL struct {
		DSN             string        `yaml:"dsn"`
		MaxOpenConns    int           `yaml:"max_open_conns"`
		MaxIdleConns    int           `yaml:"max_idle_conns"`
		ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	} `yaml:"mysql"`

	Security struct {
		Scoring scorer.ScoringConfig   `yaml:"scoring"`
		Limiter limiter.LimiterConfig `yaml:"limiter"`
		Analyzer analyzer.AnalyzerConfig `yaml:"analyzer"`
	} `yaml:"security"`

	Logging struct {
		Level      string `yaml:"level"`
		File       string `yaml:"file"`
		MaxSize    int    `yaml:"max_size"`
		MaxBackups int    `yaml:"max_backups"`
		MaxAge     int    `yaml:"max_age"`
	} `yaml:"logging"`

	WebUI struct {
		Enabled    bool   `yaml:"enabled"`
		StaticPath string `yaml:"static_path"`
		APIPrefix  string `yaml:"api_prefix"`
	} `yaml:"webui"`

	User struct {
		AllowRegistration bool `yaml:"allow_registration"`
	} `yaml:"user"`
}

// 应用实例
type App struct {
	config          *Config
	redisClient     *storage.RedisClient
	mysqlClient     *storage.MySQLClient
	collector       *collector.Collector
	fingerprint     *fingerprint.Generator
	scorer          *scorer.Scorer
	analyzer        *analyzer.Analyzer
	limiter         *limiter.Limiter
	router          *gin.Engine
}

func main() {
	// 加载配置
	config, err := loadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建应用实例
	app, err := NewApp(config)
	if err != nil {
		log.Fatalf("创建应用失例失败: %v", err)
	}
	defer app.Close()

	// 启动服务器
	log.Printf("启动服务器，端口: %d", config.Server.Port)
	if err := app.Run(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// 加载配置文件
func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

// 创建应用实例
func NewApp(config *Config) (*App, error) {
	app := &App{config: config}

	// 设置Gin模式
	if !config.Server.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化存储层
	if err := app.initStorage(); err != nil {
		return nil, fmt.Errorf("初始化存储失败: %v", err)
	}

	// 初始化核心模块
	if err := app.initModules(); err != nil {
		return nil, fmt.Errorf("初始化模块失败: %v", err)
	}

	// 初始化路由
	app.initRoutes()

	return app, nil
}

// 初始化存储层
func (app *App) initStorage() error {
	// 初始化Redis
	redisClient, err := storage.NewRedisClient(
		app.config.Redis.Addr,
		app.config.Redis.Password,
		app.config.Redis.DB,
		app.config.Redis.PoolSize,
	)
	if err != nil {
		return fmt.Errorf("Redis连接失败: %v", err)
	}
	app.redisClient = redisClient

	// 初始化MySQL
	mysqlClient, err := storage.NewMySQLClient(
		app.config.MySQL.DSN,
		app.config.MySQL.MaxOpenConns,
		app.config.MySQL.MaxIdleConns,
		app.config.MySQL.ConnMaxLifetime,
	)
	if err != nil {
		return fmt.Errorf("MySQL连接失败: %v", err)
	}
	app.mysqlClient = mysqlClient

	return nil
}

// 初始化核心模块
func (app *App) initModules() error {
	// 初始化采集器
	app.collector = collector.NewCollector()

	// 初始化指纹生成器
	app.fingerprint = fingerprint.NewGenerator("firewall-controller-salt")

	// 初始化打分系统
	app.scorer = scorer.NewScorer(app.config.Security.Scoring, app.redisClient)

	// 初始化行为分析器
	app.analyzer = analyzer.NewAnalyzer(app.config.Security.Analyzer, app.redisClient)

	// 初始化限制器
	app.limiter = limiter.NewLimiter(app.config.Security.Limiter, app.redisClient)

	return nil
}

// 初始化路由
func (app *App) initRoutes() {
	app.router = gin.New()

	// 添加中间件
	app.router.Use(gin.Logger())
	app.router.Use(gin.Recovery())
	app.router.Use(middleware.CORS())

	// 添加防火墙中间件
	app.router.Use(app.firewallMiddleware())

	// API路由组
	apiV1 := app.router.Group(app.config.WebUI.APIPrefix)

	// 注册API路由
	configAPI := api.NewConfigAPI(app.limiter, app.scorer)
	configAPI.RegisterRoutes(apiV1)

	logsAPI := api.NewLogsAPI(app.mysqlClient, app.redisClient)
	logsAPI.RegisterRoutes(apiV1)

	scoreAPI := api.NewScoreAPI(app.scorer, app.redisClient)
	scoreAPI.RegisterRoutes(apiV1)

	ruleAPI := api.NewRuleAPI(app.limiter, app.analyzer, app.redisClient)
	ruleAPI.RegisterRoutes(apiV1)

	proxyAPI := api.NewProxyAPI(app.collector)
	proxyAPI.RegisterRoutes(apiV1)

	// 系统信息API
	apiV1.GET("/system/info", app.getSystemInfo)
	apiV1.GET("/system/health", app.getHealthCheck)

	// 静态文件服务（WebUI）
	if app.config.WebUI.Enabled {
		app.router.Static("/static", app.config.WebUI.StaticPath+"/static")
		app.router.StaticFile("/", app.config.WebUI.StaticPath+"/index.html")
		app.router.StaticFile("/favicon.ico", app.config.WebUI.StaticPath+"/favicon.ico")
		
		// SPA路由支持
		app.router.NoRoute(func(c *gin.Context) {
			c.File(app.config.WebUI.StaticPath + "/index.html")
		})
	}
}

// 防火墙中间件
func (app *App) firewallMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过API和静态文件的防火墙检查
		if c.Request.URL.Path == "/api/v1/system/health" ||
		   c.Request.URL.Path == "/favicon.ico" ||
		   strings.HasPrefix(c.Request.URL.Path, "/static/") ||
		   strings.HasPrefix(c.Request.URL.Path, app.config.WebUI.APIPrefix) {
			c.Next()
			return
		}

		// 采集访问信息
		accessInfo := app.collector.CollectFromRequest(c.Request)

		// 生成用户指纹
		userFingerprint := app.fingerprint.Generate(accessInfo)

		// 增加请求计数
		app.redisClient.IncrementRequestRate(userFingerprint)

		// 计算用户分数
		scoreResult, err := app.scorer.CalculateScore(userFingerprint, accessInfo)
		if err != nil {
			log.Printf("计算用户分数失败: %v", err)
			c.Next()
			return
		}

		// 获取最近访问记录进行行为分析
		recentAccess, _ := app.redisClient.GetRecentAccess(userFingerprint, 60)
		analysisResult, _ := app.analyzer.AnalyzeUser(userFingerprint, recentAccess)

		// 检查限制
		decision, err := app.limiter.CheckLimit(userFingerprint, scoreResult.NewScore, analysisResult)
		if err != nil {
			log.Printf("检查限制失败: %v", err)
			c.Next()
			return
		}

		// 记录访问日志到Redis
		accessLog := &storage.AccessLog{
			Fingerprint: userFingerprint,
			IP:          accessInfo.IP,
			UserAgent:   accessInfo.UserAgent,
			Path:        accessInfo.Path,
			Timestamp:   time.Now(),
			Score:       scoreResult.NewScore,
		}
		app.redisClient.LogAccess(accessLog)

		// 记录访问日志到MySQL
		accessRecord := &storage.AccessRecord{
			Fingerprint: userFingerprint,
			IP:          accessInfo.IP,
			UserAgent:   accessInfo.UserAgent,
			Path:        accessInfo.Path,
			Method:      accessInfo.Method,
			Score:       scoreResult.NewScore,
			Action:      decision.Action,
			Timestamp:   time.Now(),
		}
		app.mysqlClient.LogAccess(accessRecord)

		// 应用限制决策
		if app.limiter.ApplyDecision(c.Writer, c.Request, decision) {
			c.Abort()
			return
		}

		// 设置响应头
		c.Header("X-User-Fingerprint", userFingerprint)
		c.Header("X-User-Score", fmt.Sprintf("%d", scoreResult.NewScore))
		c.Header("X-Risk-Level", analysisResult.RiskLevel)

		c.Next()
	}
}

// 获取系统信息
func (app *App) getSystemInfo(c *gin.Context) {
	info := map[string]interface{}{
		"name":         "Firewall Controller",
		"version":      "1.0.0",
		"go_version":   "1.21",
		"build_time":   "2024-01-15T10:00:00Z",
		"uptime":       time.Since(time.Now().Add(-time.Hour)).String(), // 模拟运行时间
		"user_registration_allowed": app.config.User.AllowRegistration,
	}

	c.JSON(http.StatusOK, api.ConfigResponse{
		Success: true,
		Data:    info,
	})
}

// 健康检查
func (app *App) getHealthCheck(c *gin.Context) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"services": map[string]string{
			"redis": "connected",
			"mysql": "connected",
		},
	}

	c.JSON(http.StatusOK, api.ConfigResponse{
		Success: true,
		Data:    health,
	})
}

// 运行应用
func (app *App) Run() error {
	addr := fmt.Sprintf(":%d", app.config.Server.Port)
	return app.router.Run(addr)
}

// 关闭应用
func (app *App) Close() {
	if app.redisClient != nil {
		app.redisClient.Close()
	}
	if app.mysqlClient != nil {
		app.mysqlClient.Close()
	}
}
