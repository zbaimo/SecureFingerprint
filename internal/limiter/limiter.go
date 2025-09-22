package limiter

import (
	"fmt"
	"net/http"
	"time"

	"securefingerprint/internal/analyzer"
	"securefingerprint/internal/storage"
)

// 限制器配置
type LimiterConfig struct {
	RateLimitWindow       time.Duration `yaml:"rate_limit_window"`        // 限速时间窗口
	MaxRequestsPerWindow  int           `yaml:"max_requests_per_window"`  // 每个窗口最大请求数
	BanDuration          time.Duration `yaml:"ban_duration"`             // 封禁时长
	DelayResponseMs      int           `yaml:"delay_response_ms"`        // 限速延迟时间
	WarningThreshold     int           `yaml:"warning_threshold"`        // 警告阈值
	CriticalThreshold    int           `yaml:"critical_threshold"`       // 严重阈值
}

// 默认限制器配置
var DefaultLimiterConfig = LimiterConfig{
	RateLimitWindow:      time.Minute,
	MaxRequestsPerWindow: 100,
	BanDuration:         time.Hour,
	DelayResponseMs:     1000,
	WarningThreshold:    30,  // 分数低于30时警告
	CriticalThreshold:   10,  // 分数低于10时严格限制
}

// 限制决策
type LimitDecision struct {
	Action      string        `json:"action"`       // "allow", "delay", "challenge", "ban"
	Reason      string        `json:"reason"`       // 限制原因
	Delay       time.Duration `json:"delay"`        // 延迟时间
	BanDuration time.Duration `json:"ban_duration"` // 封禁时长
	Headers     map[string]string `json:"headers"`  // 响应头
	StatusCode  int           `json:"status_code"`  // HTTP状态码
	Message     string        `json:"message"`      // 响应消息
}

type Limiter struct {
	config      LimiterConfig
	redisClient *storage.RedisClient
}

func NewLimiter(config LimiterConfig, redisClient *storage.RedisClient) *Limiter {
	return &Limiter{
		config:      config,
		redisClient: redisClient,
	}
}

// 检查并应用限制
func (l *Limiter) CheckLimit(fingerprint string, userScore int, analysisResult *analyzer.AnalysisResult) (*LimitDecision, error) {
	// 1. 首先检查是否已被封禁
	if banned, duration, err := l.redisClient.IsUserBanned(fingerprint); err == nil && banned {
		return &LimitDecision{
			Action:      "ban",
			Reason:      "用户已被封禁",
			BanDuration: duration,
			StatusCode:  403,
			Message:     fmt.Sprintf("您已被封禁，剩余时间: %v", duration.Round(time.Minute)),
			Headers: map[string]string{
				"X-Rate-Limit-Status": "banned",
				"Retry-After":         fmt.Sprintf("%.0f", duration.Seconds()),
			},
		}, nil
	}

	// 2. 检查请求频率
	if decision := l.checkRateLimit(fingerprint); decision != nil {
		return decision, nil
	}

	// 3. 基于用户分数决策
	if decision := l.checkScoreBasedLimit(fingerprint, userScore); decision != nil {
		return decision, nil
	}

	// 4. 基于行为分析结果决策
	if analysisResult != nil {
		if decision := l.checkAnalysisBasedLimit(fingerprint, analysisResult); decision != nil {
			return decision, nil
		}
	}

	// 5. 默认允许
	return &LimitDecision{
		Action:     "allow",
		Reason:     "正常访问",
		StatusCode: 200,
		Headers: map[string]string{
			"X-Rate-Limit-Status": "ok",
		},
	}, nil
}

// 检查频率限制
func (l *Limiter) checkRateLimit(fingerprint string) *LimitDecision {
	rate, err := l.redisClient.GetRequestRate(fingerprint)
	if err != nil {
		return nil
	}

	if rate > l.config.MaxRequestsPerWindow {
		// 超过频率限制，应用延迟
		delay := time.Duration(l.config.DelayResponseMs) * time.Millisecond
		
		// 根据超出程度调整延迟
		if rate > l.config.MaxRequestsPerWindow*2 {
			delay *= 3
		} else if rate > l.config.MaxRequestsPerWindow*1.5 {
			delay *= 2
		}

		return &LimitDecision{
			Action:     "delay",
			Reason:     fmt.Sprintf("请求频率过高: %d/%s", rate, l.config.RateLimitWindow),
			Delay:      delay,
			StatusCode: 429,
			Headers: map[string]string{
				"X-Rate-Limit-Status":    "rate_limited",
				"X-Rate-Limit-Limit":     fmt.Sprintf("%d", l.config.MaxRequestsPerWindow),
				"X-Rate-Limit-Remaining": "0",
				"X-Rate-Limit-Reset":     fmt.Sprintf("%d", time.Now().Add(l.config.RateLimitWindow).Unix()),
				"Retry-After":            fmt.Sprintf("%.0f", delay.Seconds()),
			},
		}
	}

	return nil
}

// 基于分数的限制检查
func (l *Limiter) checkScoreBasedLimit(fingerprint string, score int) *LimitDecision {
	if score <= 0 {
		// 分数为0或负数，封禁
		return l.banUser(fingerprint, "用户分数过低", l.config.BanDuration)
	}

	if score < l.config.CriticalThreshold {
		// 分数过低，需要人机验证
		return &LimitDecision{
			Action:     "challenge",
			Reason:     fmt.Sprintf("用户分数过低: %d", score),
			StatusCode: 429,
			Headers: map[string]string{
				"X-Rate-Limit-Status": "challenge_required",
				"X-User-Score":        fmt.Sprintf("%d", score),
			},
			Message: "需要完成人机验证",
		}
	}

	if score < l.config.WarningThreshold {
		// 分数较低，限速
		delay := time.Duration(l.config.DelayResponseMs*2) * time.Millisecond
		return &LimitDecision{
			Action:     "delay",
			Reason:     fmt.Sprintf("用户分数较低: %d", score),
			Delay:      delay,
			StatusCode: 200,
			Headers: map[string]string{
				"X-Rate-Limit-Status": "score_limited",
				"X-User-Score":        fmt.Sprintf("%d", score),
			},
		}
	}

	return nil
}

// 基于行为分析的限制检查
func (l *Limiter) checkAnalysisBasedLimit(fingerprint string, result *analyzer.AnalysisResult) *LimitDecision {
	switch result.RiskLevel {
	case "critical":
		// 严重风险，立即封禁
		duration := l.config.BanDuration * 2 // 加倍封禁时间
		return l.banUser(fingerprint, fmt.Sprintf("严重风险行为: %.1f", result.RiskScore), duration)

	case "high":
		// 高风险，需要人机验证
		return &LimitDecision{
			Action:     "challenge",
			Reason:     fmt.Sprintf("高风险行为: %.1f", result.RiskScore),
			StatusCode: 429,
			Headers: map[string]string{
				"X-Rate-Limit-Status": "high_risk",
				"X-Risk-Score":        fmt.Sprintf("%.1f", result.RiskScore),
			},
			Message: "检测到高风险行为，需要完成验证",
		}

	case "medium":
		// 中等风险，限速
		delay := time.Duration(l.config.DelayResponseMs*3) * time.Millisecond
		return &LimitDecision{
			Action:     "delay",
			Reason:     fmt.Sprintf("中等风险行为: %.1f", result.RiskScore),
			Delay:      delay,
			StatusCode: 200,
			Headers: map[string]string{
				"X-Rate-Limit-Status": "medium_risk",
				"X-Risk-Score":        fmt.Sprintf("%.1f", result.RiskScore),
			},
		}
	}

	// 检查特定行为模式
	for _, behavior := range result.Behaviors {
		if behavior.Type == "bot_behavior" && behavior.Confidence > 0.8 {
			return l.banUser(fingerprint, "检测到机器人行为", l.config.BanDuration)
		}

		if behavior.Type == "scanning_behavior" && behavior.Severity == "danger" {
			return l.banUser(fingerprint, "检测到恶意扫描", l.config.BanDuration*3)
		}
	}

	return nil
}

// 封禁用户
func (l *Limiter) banUser(fingerprint, reason string, duration time.Duration) *LimitDecision {
	// 在Redis中记录封禁
	err := l.redisClient.BanUser(fingerprint, duration)
	if err != nil {
		// 记录错误但继续执行
		fmt.Printf("封禁用户时出错: %v\n", err)
	}

	return &LimitDecision{
		Action:      "ban",
		Reason:      reason,
		BanDuration: duration,
		StatusCode:  403,
		Message:     fmt.Sprintf("您已被封禁，原因: %s，时长: %v", reason, duration.Round(time.Minute)),
		Headers: map[string]string{
			"X-Rate-Limit-Status": "banned",
			"X-Ban-Reason":        reason,
			"Retry-After":         fmt.Sprintf("%.0f", duration.Seconds()),
		},
	}
}

// 应用限制决策到HTTP响应
func (l *Limiter) ApplyDecision(w http.ResponseWriter, r *http.Request, decision *LimitDecision) bool {
	// 设置响应头
	for key, value := range decision.Headers {
		w.Header().Set(key, value)
	}

	switch decision.Action {
	case "allow":
		return false // 不阻止请求

	case "delay":
		// 延迟响应
		if decision.Delay > 0 {
			time.Sleep(decision.Delay)
		}
		return false // 延迟后允许请求

	case "challenge":
		// 返回人机验证页面
		w.WriteHeader(decision.StatusCode)
		l.writeChallengeResponse(w, decision)
		return true // 阻止请求

	case "ban":
		// 返回封禁信息
		w.WriteHeader(decision.StatusCode)
		l.writeBanResponse(w, decision)
		return true // 阻止请求

	default:
		return false
	}
}

// 写入人机验证响应
func (l *Limiter) writeChallengeResponse(w http.ResponseWriter, decision *LimitDecision) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	
	response := map[string]interface{}{
		"error":   "challenge_required",
		"message": decision.Message,
		"reason":  decision.Reason,
		"challenge": map[string]interface{}{
			"type": "captcha",
			"url":  "/api/v1/challenge",
		},
	}

	// 这里应该返回实际的验证码页面或API响应
	fmt.Fprintf(w, `{
		"error": "challenge_required",
		"message": "%s",
		"reason": "%s",
		"challenge": {
			"type": "captcha",
			"url": "/api/v1/challenge"
		}
	}`, decision.Message, decision.Reason)
}

// 写入封禁响应
func (l *Limiter) writeBanResponse(w http.ResponseWriter, decision *LimitDecision) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	
	fmt.Fprintf(w, `{
		"error": "banned",
		"message": "%s",
		"reason": "%s",
		"ban_duration": "%s",
		"retry_after": %.0f
	}`, decision.Message, decision.Reason, 
		decision.BanDuration.String(), decision.BanDuration.Seconds())
}

// 手动封禁用户
func (l *Limiter) ManualBan(fingerprint, reason string, duration time.Duration) error {
	return l.redisClient.BanUser(fingerprint, duration)
}

// 解除封禁
func (l *Limiter) Unban(fingerprint string) error {
	return l.redisClient.UnbanUser(fingerprint)
}

// 获取封禁状态
func (l *Limiter) GetBanStatus(fingerprint string) (bool, time.Duration, error) {
	return l.redisClient.IsUserBanned(fingerprint)
}

// 更新配置
func (l *Limiter) UpdateConfig(config LimiterConfig) {
	l.config = config
}

// 获取限制统计信息
func (l *Limiter) GetLimitStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 这里应该从Redis或数据库中获取实际统计数据
	// 为简化实现，返回模拟数据
	stats["total_requests"] = 15000
	stats["allowed_requests"] = 14200
	stats["delayed_requests"] = 500
	stats["challenged_requests"] = 200
	stats["banned_requests"] = 100
	stats["active_bans"] = 25
	
	return stats, nil
}

// 清理过期数据
func (l *Limiter) CleanupExpiredData() error {
	// 这里可以实现清理过期封禁记录、频率计数等
	// Redis的TTL机制会自动清理大部分数据
	// 这个方法主要用于清理一些不会自动过期的数据
	
	return nil
}

// 中间件：集成到HTTP服务器
func (l *Limiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 这里需要从请求中获取用户指纹和分数
			// 实际实现中需要与其他模块集成
			
			// 示例：从header中获取指纹
			fingerprint := r.Header.Get("X-User-Fingerprint")
			if fingerprint == "" {
				// 如果没有指纹，可能需要先生成
				next.ServeHTTP(w, r)
				return
			}

			// 增加请求计数
			l.redisClient.IncrementRequestRate(fingerprint)

			// 获取用户分数（这里需要与scorer模块集成）
			userScore, _ := l.redisClient.GetUserScore(fingerprint)
			
			// 检查限制
			decision, err := l.CheckLimit(fingerprint, userScore.Score, nil)
			if err != nil {
				// 错误处理
				http.Error(w, "Internal Server Error", 500)
				return
			}

			// 应用限制决策
			if l.ApplyDecision(w, r, decision) {
				return // 请求被阻止
			}

			// 继续处理请求
			next.ServeHTTP(w, r)
		})
	}
}

// 验证码验证
func (l *Limiter) VerifyChallenge(fingerprint, challengeResponse string) bool {
	// 这里应该实现实际的验证码验证逻辑
	// 为简化实现，假设验证总是成功
	
	// 验证成功后，可以临时提升用户分数或给予通行证
	return true
}

// 创建白名单
func (l *Limiter) AddToWhitelist(fingerprint string, duration time.Duration) error {
	key := fmt.Sprintf("whitelist:%s", fingerprint)
	return l.redisClient.client.Set(l.redisClient.ctx, key, "whitelisted", duration).Err()
}

// 检查白名单
func (l *Limiter) IsWhitelisted(fingerprint string) (bool, error) {
	key := fmt.Sprintf("whitelist:%s", fingerprint)
	_, err := l.redisClient.client.Get(l.redisClient.ctx, key).Result()
	if err != nil {
		return false, nil // 不在白名单中
	}
	return true, nil
}
