package scorer

import (
	"fmt"
	"strings"
	"time"

	"firewall-controller/internal/collector"
	"firewall-controller/internal/storage"
)

// 打分规则配置
type ScoringConfig struct {
	InitialScore           int     `yaml:"initial_score"`            // 初始分数
	NormalAccessBonus      int     `yaml:"normal_access_bonus"`      // 正常访问加分
	MaxScore              int     `yaml:"max_score"`                // 最大分数
	FrequentRequestPenalty int     `yaml:"frequent_request_penalty"` // 频繁请求扣分
	SuspiciousUAPenalty   int     `yaml:"suspicious_ua_penalty"`    // 可疑UA扣分
	BanThreshold          int     `yaml:"ban_threshold"`            // 封禁阈值
	BotPenalty            int     `yaml:"bot_penalty"`              // 机器人扣分
	ProxyPenalty          int     `yaml:"proxy_penalty"`            // 代理访问扣分
	PathSpamPenalty       int     `yaml:"path_spam_penalty"`        // 路径垃圾信息扣分
	NoRefererPenalty      int     `yaml:"no_referer_penalty"`       // 无来源扣分
}

// 默认打分配置
var DefaultScoringConfig = ScoringConfig{
	InitialScore:           100,
	NormalAccessBonus:      1,
	MaxScore:              100,
	FrequentRequestPenalty: -10,
	SuspiciousUAPenalty:   -20,
	BanThreshold:          0,
	BotPenalty:            -15,
	ProxyPenalty:          -5,
	PathSpamPenalty:       -8,
	NoRefererPenalty:      -2,
}

// 打分结果
type ScoreResult struct {
	OldScore    int                    `json:"old_score"`
	NewScore    int                    `json:"new_score"`
	Change      int                    `json:"change"`
	Reasons     []string               `json:"reasons"`
	Action      string                 `json:"action"` // "allow", "limit", "ban"
	Details     map[string]interface{} `json:"details"`
	Timestamp   time.Time              `json:"timestamp"`
}

type Scorer struct {
	config      ScoringConfig
	redisClient *storage.RedisClient
}

func NewScorer(config ScoringConfig, redisClient *storage.RedisClient) *Scorer {
	return &Scorer{
		config:      config,
		redisClient: redisClient,
	}
}

// 计算访问分数
func (s *Scorer) CalculateScore(fingerprint string, info *collector.AccessInfo) (*ScoreResult, error) {
	// 获取当前用户分数
	userScore, err := s.redisClient.GetUserScore(fingerprint)
	if err != nil {
		return nil, fmt.Errorf("获取用户分数失败: %v", err)
	}

	oldScore := userScore.Score
	newScore := oldScore
	var reasons []string
	details := make(map[string]interface{})

	// 基础分数调整
	scoreAdjustments := s.analyzeAccess(info, userScore)
	
	for _, adjustment := range scoreAdjustments {
		newScore += adjustment.Points
		reasons = append(reasons, adjustment.Reason)
		details[adjustment.Category] = adjustment.Points
	}

	// 确保分数在合理范围内
	if newScore > s.config.MaxScore {
		newScore = s.config.MaxScore
	}
	if newScore < -50 { // 设置最低分数限制
		newScore = -50
	}

	// 更新用户分数
	userScore.Score = newScore
	userScore.LastSeen = time.Now()
	userScore.RequestCount++
	
	err = s.redisClient.UpdateUserScore(fingerprint, userScore)
	if err != nil {
		return nil, fmt.Errorf("更新用户分数失败: %v", err)
	}

	// 确定动作
	action := s.determineAction(newScore, info)

	result := &ScoreResult{
		OldScore:  oldScore,
		NewScore:  newScore,
		Change:    newScore - oldScore,
		Reasons:   reasons,
		Action:    action,
		Details:   details,
		Timestamp: time.Now(),
	}

	return result, nil
}

// 分析访问行为并返回分数调整
func (s *Scorer) analyzeAccess(info *collector.AccessInfo, userScore *storage.UserScore) []ScoreAdjustment {
	var adjustments []ScoreAdjustment

	// 1. 检查是否为机器人
	if info.IsBot {
		adjustments = append(adjustments, ScoreAdjustment{
			Points:   s.config.BotPenalty,
			Reason:   "检测到机器人行为",
			Category: "bot_detection",
		})
	} else {
		// 正常访问加分
		adjustments = append(adjustments, ScoreAdjustment{
			Points:   s.config.NormalAccessBonus,
			Reason:   "正常用户访问",
			Category: "normal_access",
		})
	}

	// 2. 检查User-Agent可疑性
	if s.isSuspiciousUserAgent(info.UserAgent) {
		adjustments = append(adjustments, ScoreAdjustment{
			Points:   s.config.SuspiciousUAPenalty,
			Reason:   "可疑的User-Agent",
			Category: "suspicious_ua",
		})
	}

	// 3. 检查网络类型
	if info.NetworkType == "proxy" {
		adjustments = append(adjustments, ScoreAdjustment{
			Points:   s.config.ProxyPenalty,
			Reason:   "通过代理访问",
			Category: "proxy_access",
		})
	}

	// 4. 检查访问路径
	if s.isSuspiciousPath(info.Path) {
		adjustments = append(adjustments, ScoreAdjustment{
			Points:   s.config.PathSpamPenalty,
			Reason:   "访问可疑路径",
			Category: "suspicious_path",
		})
	}

	// 5. 检查Referer
	if info.Referer == "" && info.Method == "GET" {
		adjustments = append(adjustments, ScoreAdjustment{
			Points:   s.config.NoRefererPenalty,
			Reason:   "缺少来源信息",
			Category: "no_referer",
		})
	}

	// 6. 检查请求频率（需要查询Redis）
	if s.redisClient != nil {
		if rate, err := s.redisClient.GetRequestRate(info.IP); err == nil && rate > 50 {
			penalty := s.config.FrequentRequestPenalty
			// 根据频率调整扣分力度
			if rate > 100 {
				penalty *= 2
			}
			adjustments = append(adjustments, ScoreAdjustment{
				Points:   penalty,
				Reason:   fmt.Sprintf("请求过于频繁 (%d/分钟)", rate),
				Category: "frequent_requests",
			})
		}
	}

	return adjustments
}

// 分数调整结构
type ScoreAdjustment struct {
	Points   int    `json:"points"`
	Reason   string `json:"reason"`
	Category string `json:"category"`
}

// 判断User-Agent是否可疑
func (s *Scorer) isSuspiciousUserAgent(userAgent string) bool {
	if userAgent == "" {
		return true
	}

	ua := strings.ToLower(userAgent)
	
	// 可疑模式
	suspiciousPatterns := []string{
		"curl", "wget", "python", "java", "go-http",
		"postman", "insomnia", "httpie",
		"scanner", "test", "bot", "crawler",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(ua, pattern) {
			return true
		}
	}

	// 检查User-Agent长度
	if len(userAgent) < 20 {
		return true
	}

	// 检查是否包含版本信息（正常浏览器通常包含版本）
	hasVersion := strings.Contains(ua, "version/") || 
		strings.Contains(ua, "chrome/") ||
		strings.Contains(ua, "firefox/") ||
		strings.Contains(ua, "safari/")
	
	if !hasVersion {
		return true
	}

	return false
}

// 判断访问路径是否可疑
func (s *Scorer) isSuspiciousPath(path string) bool {
	if path == "" {
		return false
	}

	path = strings.ToLower(path)
	
	// 可疑路径模式
	suspiciousPaths := []string{
		"/admin", "/wp-admin", "/phpmyadmin",
		"/.env", "/.git", "/config",
		"/api/v1/users", "/api/admin",
		"/xmlrpc.php", "/wp-login.php",
		"/../", "/./", // 路径遍历
		"<script", "javascript:", // XSS尝试
		"union select", "drop table", // SQL注入尝试
	}

	for _, suspicious := range suspiciousPaths {
		if strings.Contains(path, suspicious) {
			return true
		}
	}

	// 检查是否包含过多的特殊字符
	specialChars := strings.Count(path, "%") + strings.Count(path, "&") + 
		strings.Count(path, "=") + strings.Count(path, "?")
	
	if specialChars > 10 {
		return true
	}

	return false
}

// 确定应该采取的行动
func (s *Scorer) determineAction(score int, info *collector.AccessInfo) string {
	// 分数太低，封禁
	if score <= s.config.BanThreshold {
		return "ban"
	}

	// 分数较低或检测到机器人，限制
	if score < 30 || info.IsBot {
		return "limit"
	}

	// 代理访问，轻度限制
	if info.NetworkType == "proxy" && score < 70 {
		return "limit"
	}

	return "allow"
}

// 获取用户分数历史趋势
func (s *Scorer) GetScoreTrend(fingerprint string, hours int) ([]ScoreTrendPoint, error) {
	// 这里应该从数据库或缓存中获取历史分数数据
	// 为简化实现，返回模拟数据
	var trend []ScoreTrendPoint
	
	now := time.Now()
	for i := hours; i >= 0; i-- {
		point := ScoreTrendPoint{
			Timestamp: now.Add(-time.Duration(i) * time.Hour),
			Score:     85 + (i%10 - 5), // 模拟数据
		}
		trend = append(trend, point)
	}
	
	return trend, nil
}

// 分数趋势点
type ScoreTrendPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Score     int       `json:"score"`
}

// 批量更新分数（用于定时任务）
func (s *Scorer) BatchUpdateScores() error {
	// 这里可以实现批量分数衰减或恢复逻辑
	// 例如：每天为所有用户恢复一定分数
	
	// 实现示例：为所有用户每日恢复5分
	// 这需要遍历所有用户，在实际实现中应该通过数据库批量操作
	
	return nil
}

// 重置用户分数
func (s *Scorer) ResetUserScore(fingerprint string) error {
	userScore := &storage.UserScore{
		Score:        s.config.InitialScore,
		LastSeen:     time.Now(),
		RequestCount: 0,
	}
	
	return s.redisClient.UpdateUserScore(fingerprint, userScore)
}

// 获取分数统计信息
func (s *Scorer) GetScoreStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 这里应该从数据库中查询实际统计数据
	// 为简化实现，返回模拟数据
	stats["total_users"] = 1500
	stats["high_score_users"] = 1200  // 分数 > 80
	stats["medium_score_users"] = 250 // 分数 50-80
	stats["low_score_users"] = 50     // 分数 < 50
	stats["banned_users"] = 15
	
	return stats, nil
}
