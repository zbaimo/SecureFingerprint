package analyzer

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"firewall-controller/internal/collector"
	"firewall-controller/internal/storage"
)

// 行为分析配置
type AnalyzerConfig struct {
	SuspiciousRequestThreshold int           `yaml:"suspicious_request_threshold"` // 可疑请求阈值
	PathRepeatThreshold       int           `yaml:"path_repeat_threshold"`        // 路径重复阈值
	BotDetectionEnabled       bool          `yaml:"bot_detection_enabled"`        // 启用机器人检测
	AnalysisWindow           time.Duration `yaml:"analysis_window"`              // 分析时间窗口
	PatternDetectionEnabled   bool          `yaml:"pattern_detection_enabled"`    // 启用模式检测
}

// 默认分析配置
var DefaultAnalyzerConfig = AnalyzerConfig{
	SuspiciousRequestThreshold: 50,
	PathRepeatThreshold:       10,
	BotDetectionEnabled:       true,
	AnalysisWindow:           time.Hour,
	PatternDetectionEnabled:   true,
}

// 行为分析结果
type AnalysisResult struct {
	Fingerprint     string                 `json:"fingerprint"`
	RiskLevel       string                 `json:"risk_level"`       // "low", "medium", "high", "critical"
	RiskScore       float64                `json:"risk_score"`       // 0-100
	Behaviors       []DetectedBehavior     `json:"behaviors"`
	Recommendations []string               `json:"recommendations"`
	Details         map[string]interface{} `json:"details"`
	Timestamp       time.Time              `json:"timestamp"`
}

// 检测到的行为
type DetectedBehavior struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"` // "info", "warning", "danger"
	Description string    `json:"description"`
	Evidence    []string  `json:"evidence"`
	Confidence  float64   `json:"confidence"` // 0-1
	Timestamp   time.Time `json:"timestamp"`
}

// 访问模式
type AccessPattern struct {
	PathFrequency    map[string]int    `json:"path_frequency"`
	TimeDistribution map[int]int       `json:"time_distribution"` // 按小时统计
	RequestRate      []RatePoint       `json:"request_rate"`
	UserAgents       map[string]int    `json:"user_agents"`
	Methods          map[string]int    `json:"methods"`
}

type RatePoint struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int       `json:"count"`
}

type Analyzer struct {
	config      AnalyzerConfig
	redisClient *storage.RedisClient
}

func NewAnalyzer(config AnalyzerConfig, redisClient *storage.RedisClient) *Analyzer {
	return &Analyzer{
		config:      config,
		redisClient: redisClient,
	}
}

// 分析用户行为
func (a *Analyzer) AnalyzeUser(fingerprint string, recentAccess []storage.AccessLog) (*AnalysisResult, error) {
	if len(recentAccess) == 0 {
		return &AnalysisResult{
			Fingerprint: fingerprint,
			RiskLevel:   "low",
			RiskScore:   0,
			Timestamp:   time.Now(),
		}, nil
	}

	// 提取访问模式
	pattern := a.extractAccessPattern(recentAccess)
	
	// 检测各种行为
	behaviors := a.detectBehaviors(pattern, recentAccess)
	
	// 计算风险分数
	riskScore := a.calculateRiskScore(behaviors, pattern)
	
	// 确定风险等级
	riskLevel := a.determineRiskLevel(riskScore)
	
	// 生成建议
	recommendations := a.generateRecommendations(behaviors, riskScore)

	result := &AnalysisResult{
		Fingerprint:     fingerprint,
		RiskLevel:       riskLevel,
		RiskScore:       riskScore,
		Behaviors:       behaviors,
		Recommendations: recommendations,
		Details: map[string]interface{}{
			"total_requests":    len(recentAccess),
			"unique_paths":      len(pattern.PathFrequency),
			"unique_user_agents": len(pattern.UserAgents),
			"analysis_window":   a.config.AnalysisWindow.String(),
		},
		Timestamp: time.Now(),
	}

	return result, nil
}

// 提取访问模式
func (a *Analyzer) extractAccessPattern(logs []storage.AccessLog) *AccessPattern {
	pattern := &AccessPattern{
		PathFrequency:    make(map[string]int),
		TimeDistribution: make(map[int]int),
		UserAgents:       make(map[string]int),
		Methods:          make(map[string]int),
	}

	// 按时间排序
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp.Before(logs[j].Timestamp)
	})

	// 统计各种模式
	for _, log := range logs {
		// 路径频率
		pattern.PathFrequency[log.Path]++
		
		// 时间分布（按小时）
		hour := log.Timestamp.Hour()
		pattern.TimeDistribution[hour]++
		
		// User-Agent统计
		pattern.UserAgents[log.UserAgent]++
		
		// HTTP方法统计
		if log.Method != "" {
			pattern.Methods[log.Method]++
		}
	}

	// 计算请求频率
	pattern.RequestRate = a.calculateRequestRate(logs)

	return pattern
}

// 计算请求频率
func (a *Analyzer) calculateRequestRate(logs []storage.AccessLog) []RatePoint {
	if len(logs) == 0 {
		return nil
	}

	// 按分钟统计请求数
	minuteCounts := make(map[int64]int)
	
	for _, log := range logs {
		minute := log.Timestamp.Unix() / 60 // 转换为分钟时间戳
		minuteCounts[minute]++
	}

	// 转换为RatePoint数组
	var ratePoints []RatePoint
	for minute, count := range minuteCounts {
		ratePoints = append(ratePoints, RatePoint{
			Timestamp: time.Unix(minute*60, 0),
			Count:     count,
		})
	}

	// 按时间排序
	sort.Slice(ratePoints, func(i, j int) bool {
		return ratePoints[i].Timestamp.Before(ratePoints[j].Timestamp)
	})

	return ratePoints
}

// 检测行为模式
func (a *Analyzer) detectBehaviors(pattern *AccessPattern, logs []storage.AccessLog) []DetectedBehavior {
	var behaviors []DetectedBehavior

	// 1. 检测频繁请求
	if behavior := a.detectFrequentRequests(pattern); behavior != nil {
		behaviors = append(behaviors, *behavior)
	}

	// 2. 检测路径垃圾信息
	if behavior := a.detectPathSpam(pattern); behavior != nil {
		behaviors = append(behaviors, *behavior)
	}

	// 3. 检测机器人行为
	if a.config.BotDetectionEnabled {
		if behavior := a.detectBotBehavior(pattern, logs); behavior != nil {
			behaviors = append(behaviors, *behavior)
		}
	}

	// 4. 检测扫描行为
	if behavior := a.detectScanningBehavior(pattern); behavior != nil {
		behaviors = append(behaviors, *behavior)
	}

	// 5. 检测异常时间模式
	if behavior := a.detectAbnormalTimePattern(pattern); behavior != nil {
		behaviors = append(behaviors, *behavior)
	}

	// 6. 检测User-Agent异常
	if behavior := a.detectUserAgentAnomalies(pattern); behavior != nil {
		behaviors = append(behaviors, *behavior)
	}

	return behaviors
}

// 检测频繁请求
func (a *Analyzer) detectFrequentRequests(pattern *AccessPattern) *DetectedBehavior {
	maxRate := 0
	var peakTimes []string
	
	for _, point := range pattern.RequestRate {
		if point.Count > maxRate {
			maxRate = point.Count
		}
		if point.Count > a.config.SuspiciousRequestThreshold {
			peakTimes = append(peakTimes, point.Timestamp.Format("15:04"))
		}
	}

	if maxRate > a.config.SuspiciousRequestThreshold {
		severity := "warning"
		if maxRate > a.config.SuspiciousRequestThreshold*2 {
			severity = "danger"
		}

		return &DetectedBehavior{
			Type:        "frequent_requests",
			Severity:    severity,
			Description: fmt.Sprintf("检测到频繁请求，峰值: %d请求/分钟", maxRate),
			Evidence:    []string{fmt.Sprintf("峰值时间: %s", strings.Join(peakTimes, ", "))},
			Confidence:  0.9,
			Timestamp:   time.Now(),
		}
	}

	return nil
}

// 检测路径垃圾信息
func (a *Analyzer) detectPathSpam(pattern *AccessPattern) *DetectedBehavior {
	var suspiciousPaths []string
	totalRequests := 0
	
	for path, count := range pattern.PathFrequency {
		totalRequests += count
		if count > a.config.PathRepeatThreshold {
			suspiciousPaths = append(suspiciousPaths, fmt.Sprintf("%s (%d次)", path, count))
		}
	}

	if len(suspiciousPaths) > 0 {
		pathDiversity := float64(len(pattern.PathFrequency)) / float64(totalRequests)
		
		severity := "info"
		if pathDiversity < 0.1 {
			severity = "warning"
		}
		if pathDiversity < 0.05 {
			severity = "danger"
		}

		return &DetectedBehavior{
			Type:        "path_spam",
			Severity:    severity,
			Description: fmt.Sprintf("检测到路径重复访问，路径多样性: %.2f%%", pathDiversity*100),
			Evidence:    suspiciousPaths,
			Confidence:  0.8,
			Timestamp:   time.Now(),
		}
	}

	return nil
}

// 检测机器人行为
func (a *Analyzer) detectBotBehavior(pattern *AccessPattern, logs []storage.AccessLog) *DetectedBehavior {
	var botIndicators []string
	confidence := 0.0

	// 检查User-Agent
	for ua, count := range pattern.UserAgents {
		if a.isBotUserAgent(ua) {
			botIndicators = append(botIndicators, fmt.Sprintf("机器人UA: %s (%d次)", ua, count))
			confidence += 0.3
		}
	}

	// 检查请求模式规律性
	if a.hasRegularPattern(pattern.RequestRate) {
		botIndicators = append(botIndicators, "检测到规律性请求模式")
		confidence += 0.2
	}

	// 检查缺少常见浏览器行为
	if a.lacksHumanBehavior(pattern) {
		botIndicators = append(botIndicators, "缺少人类用户常见行为")
		confidence += 0.3
	}

	if confidence > 0.5 {
		severity := "warning"
		if confidence > 0.8 {
			severity = "danger"
		}

		return &DetectedBehavior{
			Type:        "bot_behavior",
			Severity:    severity,
			Description: "检测到疑似机器人行为",
			Evidence:    botIndicators,
			Confidence:  math.Min(confidence, 1.0),
			Timestamp:   time.Now(),
		}
	}

	return nil
}

// 检测扫描行为
func (a *Analyzer) detectScanningBehavior(pattern *AccessPattern) *DetectedBehavior {
	var scanningIndicators []string
	scanningScore := 0.0

	// 检查是否访问了常见的扫描路径
	scanPaths := []string{
		"/admin", "/wp-admin", "/.env", "/.git",
		"/config", "/backup", "/test", "/api",
		"/phpmyadmin", "/xmlrpc.php",
	}

	accessedScanPaths := 0
	for path := range pattern.PathFrequency {
		pathLower := strings.ToLower(path)
		for _, scanPath := range scanPaths {
			if strings.Contains(pathLower, scanPath) {
				accessedScanPaths++
				scanningIndicators = append(scanningIndicators, fmt.Sprintf("访问扫描路径: %s", path))
				break
			}
		}
	}

	if accessedScanPaths > 0 {
		scanningScore = float64(accessedScanPaths) / float64(len(scanPaths))
		
		// 检查404错误率（需要从日志中获取状态码信息）
		// 这里简化处理，假设扫描行为会产生较多404
		
		if scanningScore > 0.2 {
			severity := "warning"
			if scanningScore > 0.5 {
				severity = "danger"
			}

			return &DetectedBehavior{
				Type:        "scanning_behavior",
				Severity:    severity,
				Description: fmt.Sprintf("检测到扫描行为，访问了%d个扫描相关路径", accessedScanPaths),
				Evidence:    scanningIndicators,
				Confidence:  scanningScore,
				Timestamp:   time.Now(),
			}
		}
	}

	return nil
}

// 检测异常时间模式
func (a *Analyzer) detectAbnormalTimePattern(pattern *AccessPattern) *DetectedBehavior {
	if len(pattern.TimeDistribution) == 0 {
		return nil
	}

	// 计算访问时间的标准差
	totalRequests := 0
	for _, count := range pattern.TimeDistribution {
		totalRequests += count
	}

	// 检查是否在非正常时间（深夜）有大量访问
	nightRequests := pattern.TimeDistribution[0] + pattern.TimeDistribution[1] + 
		pattern.TimeDistribution[2] + pattern.TimeDistribution[3] + 
		pattern.TimeDistribution[4] + pattern.TimeDistribution[5]

	nightRatio := float64(nightRequests) / float64(totalRequests)
	
	if nightRatio > 0.5 && totalRequests > 10 {
		return &DetectedBehavior{
			Type:        "abnormal_time_pattern",
			Severity:    "warning",
			Description: fmt.Sprintf("检测到异常时间访问模式，%.1f%%的访问发生在深夜", nightRatio*100),
			Evidence:    []string{fmt.Sprintf("深夜(0-6点)访问次数: %d", nightRequests)},
			Confidence:  0.6,
			Timestamp:   time.Now(),
		}
	}

	return nil
}

// 检测User-Agent异常
func (a *Analyzer) detectUserAgentAnomalies(pattern *AccessPattern) *DetectedBehavior {
	if len(pattern.UserAgents) == 0 {
		return &DetectedBehavior{
			Type:        "missing_user_agent",
			Severity:    "warning",
			Description: "缺少User-Agent信息",
			Evidence:    []string{"所有请求都没有User-Agent"},
			Confidence:  0.9,
			Timestamp:   time.Now(),
		}
	}

	// 检查User-Agent多样性
	var suspiciousUAs []string
	for ua, count := range pattern.UserAgents {
		if ua == "" {
			suspiciousUAs = append(suspiciousUAs, fmt.Sprintf("空UA (%d次)", count))
		} else if len(ua) < 20 {
			suspiciousUAs = append(suspiciousUAs, fmt.Sprintf("过短UA: %s (%d次)", ua, count))
		} else if a.isBotUserAgent(ua) {
			suspiciousUAs = append(suspiciousUAs, fmt.Sprintf("机器人UA: %s (%d次)", ua, count))
		}
	}

	if len(suspiciousUAs) > 0 {
		return &DetectedBehavior{
			Type:        "suspicious_user_agent",
			Severity:    "warning",
			Description: "检测到可疑的User-Agent",
			Evidence:    suspiciousUAs,
			Confidence:  0.7,
			Timestamp:   time.Now(),
		}
	}

	return nil
}

// 辅助函数：判断是否为机器人User-Agent
func (a *Analyzer) isBotUserAgent(ua string) bool {
	if ua == "" {
		return true
	}

	uaLower := strings.ToLower(ua)
	botKeywords := []string{
		"bot", "crawler", "spider", "scraper",
		"curl", "wget", "python", "java",
		"postman", "insomnia", "httpie",
	}

	for _, keyword := range botKeywords {
		if strings.Contains(uaLower, keyword) {
			return true
		}
	}

	return false
}

// 辅助函数：检查是否有规律性模式
func (a *Analyzer) hasRegularPattern(ratePoints []RatePoint) bool {
	if len(ratePoints) < 3 {
		return false
	}

	// 简单检查：如果请求间隔非常规律，可能是机器人
	var intervals []int64
	for i := 1; i < len(ratePoints); i++ {
		interval := ratePoints[i].Timestamp.Unix() - ratePoints[i-1].Timestamp.Unix()
		intervals = append(intervals, interval)
	}

	// 检查间隔的方差
	if len(intervals) > 0 {
		var sum int64
		for _, interval := range intervals {
			sum += interval
		}
		avg := float64(sum) / float64(len(intervals))

		var variance float64
		for _, interval := range intervals {
			variance += math.Pow(float64(interval)-avg, 2)
		}
		variance /= float64(len(intervals))

		// 如果方差很小，说明间隔很规律
		return variance < 10 && avg < 120 // 平均间隔小于2分钟且很规律
	}

	return false
}

// 辅助函数：检查是否缺少人类行为
func (a *Analyzer) lacksHumanBehavior(pattern *AccessPattern) bool {
	// 人类用户通常会：
	// 1. 访问多种类型的资源
	// 2. 有一定的页面停留时间
	// 3. 访问静态资源（CSS、JS、图片）
	
	hasStaticResources := false
	for path := range pattern.PathFrequency {
		pathLower := strings.ToLower(path)
		if strings.Contains(pathLower, ".css") || 
		   strings.Contains(pathLower, ".js") || 
		   strings.Contains(pathLower, ".png") || 
		   strings.Contains(pathLower, ".jpg") || 
		   strings.Contains(pathLower, ".ico") {
			hasStaticResources = true
			break
		}
	}

	// 如果只有API调用，没有静态资源访问，可能是机器人
	return !hasStaticResources && len(pattern.PathFrequency) > 5
}

// 计算风险分数
func (a *Analyzer) calculateRiskScore(behaviors []DetectedBehavior, pattern *AccessPattern) float64 {
	score := 0.0

	// 基于行为计算分数
	for _, behavior := range behaviors {
		switch behavior.Severity {
		case "info":
			score += 10 * behavior.Confidence
		case "warning":
			score += 30 * behavior.Confidence
		case "danger":
			score += 60 * behavior.Confidence
		}
	}

	// 基于访问模式调整分数
	if len(pattern.PathFrequency) > 100 {
		score += 20 // 访问路径过多
	}

	if len(pattern.UserAgents) > 10 {
		score += 15 // User-Agent变化频繁
	}

	// 确保分数在0-100范围内
	return math.Min(score, 100)
}

// 确定风险等级
func (a *Analyzer) determineRiskLevel(score float64) string {
	if score >= 80 {
		return "critical"
	} else if score >= 60 {
		return "high"
	} else if score >= 30 {
		return "medium"
	}
	return "low"
}

// 生成建议
func (a *Analyzer) generateRecommendations(behaviors []DetectedBehavior, riskScore float64) []string {
	var recommendations []string

	// 基于检测到的行为生成建议
	behaviorTypes := make(map[string]bool)
	for _, behavior := range behaviors {
		behaviorTypes[behavior.Type] = true
	}

	if behaviorTypes["frequent_requests"] {
		recommendations = append(recommendations, "建议启用请求频率限制")
	}

	if behaviorTypes["bot_behavior"] {
		recommendations = append(recommendations, "建议启用机器人验证（如验证码）")
	}

	if behaviorTypes["scanning_behavior"] {
		recommendations = append(recommendations, "建议封禁该用户，疑似恶意扫描")
	}

	if behaviorTypes["path_spam"] {
		recommendations = append(recommendations, "建议限制对特定路径的访问频率")
	}

	// 基于风险分数生成建议
	if riskScore >= 80 {
		recommendations = append(recommendations, "建议立即封禁该用户")
	} else if riskScore >= 60 {
		recommendations = append(recommendations, "建议对该用户进行严格限制")
	} else if riskScore >= 30 {
		recommendations = append(recommendations, "建议增加监控频率")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "用户行为正常，继续监控")
	}

	return recommendations
}
