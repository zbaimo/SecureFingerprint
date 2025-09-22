package collector

import (
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// 访问信息结构体
type AccessInfo struct {
	IP            string            `json:"ip"`
	OriginalIP    string            `json:"original_ip"`    // 原始RemoteAddr
	ProxyChain    []string          `json:"proxy_chain"`    // 代理链
	UserAgent     string            `json:"user_agent"`
	Referer       string            `json:"referer"`
	Path          string            `json:"path"`
	Method        string            `json:"method"`
	Headers       map[string]string `json:"headers"`
	ProxyHeaders  map[string]string `json:"proxy_headers"`  // 代理相关头
	DeviceType    string            `json:"device_type"`
	NetworkType   string            `json:"network_type"`
	IsBot         bool              `json:"is_bot"`
	LoginStatus   bool              `json:"login_status"`
	IsBehindProxy bool              `json:"is_behind_proxy"` // 是否通过代理
	Timestamp     time.Time         `json:"timestamp"`
}

type Collector struct {
	botPatterns    []*regexp.Regexp
	mobilePatterns []*regexp.Regexp
}

func NewCollector() *Collector {
	// 常见机器人User-Agent模式
	botPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(bot|crawler|spider|scraper|curl|wget|python|java|go-http)`),
		regexp.MustCompile(`(?i)(googlebot|bingbot|slurp|duckduckbot|baiduspider|yandexbot)`),
		regexp.MustCompile(`(?i)(facebookexternalhit|twitterbot|linkedinbot|whatsapp)`),
		regexp.MustCompile(`(?i)(postman|insomnia|httpie|apache-httpclient)`),
	}

	// 移动设备User-Agent模式
	mobilePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(mobile|android|iphone|ipad|ipod|blackberry|windows phone)`),
		regexp.MustCompile(`(?i)(opera mini|opera mobi|samsung|nokia|huawei|xiaomi)`),
	}

	return &Collector{
		botPatterns:    botPatterns,
		mobilePatterns: mobilePatterns,
	}
}

// 从HTTP请求中采集访问信息
func (c *Collector) CollectFromRequest(r *http.Request) *AccessInfo {
	// 获取原始IP
	originalHost, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		originalHost = r.RemoteAddr
	}

	info := &AccessInfo{
		IP:           c.extractIP(r),
		OriginalIP:   originalHost,
		UserAgent:    r.UserAgent(),
		Referer:      r.Referer(),
		Path:         r.URL.Path,
		Method:       r.Method,
		Headers:      c.extractHeaders(r),
		ProxyHeaders: c.extractProxyHeaders(r),
		Timestamp:    time.Now(),
	}

	// 检测是否通过代理
	info.IsBehindProxy = c.detectProxy(r)
	
	// 提取代理链
	info.ProxyChain = c.extractProxyChain(r)

	// 分析设备类型
	info.DeviceType = c.detectDeviceType(info.UserAgent)
	
	// 分析网络类型
	info.NetworkType = c.detectNetworkType(info.IP)
	
	// 检测是否为机器人
	info.IsBot = c.detectBot(info.UserAgent)
	
	// 检测登录状态（通过cookie或session）
	info.LoginStatus = c.detectLoginStatus(r)

	return info
}

// 提取真实IP地址
func (c *Collector) extractIP(r *http.Request) string {
	// 代理头优先级列表（按可信度排序）
	proxyHeaders := []string{
		"CF-Connecting-IP",    // Cloudflare
		"True-Client-IP",      // Akamai, Cloudflare
		"X-Real-IP",          // Nginx
		"X-Forwarded-For",    // 标准代理头
		"X-Client-IP",        // Apache
		"X-Forwarded",        // 非标准
		"X-Cluster-Client-IP", // 集群代理
		"Forwarded-For",      // RFC 7239
		"Forwarded",          // RFC 7239
	}

	// 按优先级检查代理头
	for _, header := range proxyHeaders {
		if ip := r.Header.Get(header); ip != "" {
			// 处理可能包含多个IP的情况
			if realIP := c.parseIPFromHeader(ip, header); realIP != "" {
				// 验证IP是否有效且不是内网IP（在代理环境中）
				if c.isValidPublicIP(realIP) {
					return realIP
				}
			}
		}
	}

	// 如果没有找到有效的代理头，使用RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// 从代理头中解析IP
func (c *Collector) parseIPFromHeader(headerValue, headerName string) string {
	headerValue = strings.TrimSpace(headerValue)
	
	switch headerName {
	case "X-Forwarded-For":
		// X-Forwarded-For: client, proxy1, proxy2
		// 取第一个IP（客户端IP）
		if ips := strings.Split(headerValue, ","); len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
		
	case "Forwarded":
		// RFC 7239: Forwarded: for=192.0.2.60;proto=http;by=203.0.113.43
		// 解析 for= 参数
		parts := strings.Split(headerValue, ";")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "for=") {
				forValue := strings.TrimPrefix(part, "for=")
				// 移除可能的引号和端口号
				forValue = strings.Trim(forValue, "\"'")
				if colonIndex := strings.LastIndex(forValue, ":"); colonIndex > 0 {
					// 检查是否是IPv6格式 [::1]:port
					if !strings.HasPrefix(forValue, "[") {
						forValue = forValue[:colonIndex]
					}
				}
				return forValue
			}
		}
		
	default:
		// 其他头直接返回，但需要清理
		// 移除可能的端口号
		if colonIndex := strings.LastIndex(headerValue, ":"); colonIndex > 0 {
			// 检查是否是IPv6格式
			if !strings.HasPrefix(headerValue, "[") && !strings.Contains(headerValue[:colonIndex], ":") {
				headerValue = headerValue[:colonIndex]
			}
		}
		return headerValue
	}
	
	return headerValue
}

// 验证是否为有效的公网IP
func (c *Collector) isValidPublicIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// 检查是否为有效IP格式
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return false
	}

	// 检查是否为内网IP
	return !c.isPrivateIP(ip)
}

// 提取关键HTTP头信息
func (c *Collector) extractHeaders(r *http.Request) map[string]string {
	headers := make(map[string]string)
	
	// 只提取关键头信息
	importantHeaders := []string{
		"Accept", "Accept-Language", "Accept-Encoding",
		"Connection", "Upgrade-Insecure-Requests",
		"Sec-Ch-Ua", "Sec-Ch-Ua-Mobile", "Sec-Ch-Ua-Platform",
		"DNT", "Sec-Fetch-Dest", "Sec-Fetch-Mode", "Sec-Fetch-Site",
	}

	for _, header := range importantHeaders {
		if value := r.Header.Get(header); value != "" {
			headers[header] = value
		}
	}

	return headers
}

// 检测设备类型
func (c *Collector) detectDeviceType(userAgent string) string {
	ua := strings.ToLower(userAgent)
	
	// 检查是否为移动设备
	for _, pattern := range c.mobilePatterns {
		if pattern.MatchString(ua) {
			if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
				return "tablet"
			}
			return "mobile"
		}
	}
	
	// 检查是否为机器人
	for _, pattern := range c.botPatterns {
		if pattern.MatchString(ua) {
			return "bot"
		}
	}
	
	return "desktop"
}

// 检测网络类型
func (c *Collector) detectNetworkType(ip string) string {
	// 解析IP地址
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "unknown"
	}

	// 检查是否为内网IP
	if c.isPrivateIP(parsedIP) {
		return "private"
	}

	// 检查是否为已知的代理/VPN IP段
	if c.isProxyIP(parsedIP) {
		return "proxy"
	}

	// 检查是否为移动网络IP段（这里需要根据实际情况配置）
	if c.isMobileIP(parsedIP) {
		return "mobile"
	}

	return "broadband"
}

// 检查是否为内网IP
func (c *Collector) isPrivateIP(ip net.IP) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
	}

	for _, cidr := range privateRanges {
		_, network, _ := net.ParseCIDR(cidr)
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// 检查是否为代理IP（简单实现，实际需要维护IP黑名单库）
func (c *Collector) isProxyIP(ip net.IP) bool {
	// 这里应该查询代理IP数据库
	// 为演示目的，只检查一些已知的公共代理段
	proxyRanges := []string{
		"8.8.8.0/24",    // 示例：Google DNS段
		"1.1.1.0/24",    // 示例：Cloudflare DNS段
	}

	for _, cidr := range proxyRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// 检查是否为移动网络IP（需要根据实际运营商IP段配置）
func (c *Collector) isMobileIP(ip net.IP) bool {
	// 这里应该维护移动运营商的IP段数据库
	// 为演示目的，返回false
	return false
}

// 检测是否为机器人
func (c *Collector) detectBot(userAgent string) bool {
	if userAgent == "" {
		return true // 没有User-Agent通常是机器人
	}

	ua := strings.ToLower(userAgent)
	
	for _, pattern := range c.botPatterns {
		if pattern.MatchString(ua) {
			return true
		}
	}

	// 检查User-Agent是否过于简单
	if len(userAgent) < 10 {
		return true
	}

	// 检查是否包含常见浏览器标识
	browserKeywords := []string{"chrome", "firefox", "safari", "edge", "opera"}
	hasValidBrowser := false
	for _, keyword := range browserKeywords {
		if strings.Contains(ua, keyword) {
			hasValidBrowser = true
			break
		}
	}

	// 如果没有浏览器标识但有明显的编程语言标识，可能是机器人
	if !hasValidBrowser {
		programmingKeywords := []string{"python", "java", "node", "php", "ruby", "go", "rust"}
		for _, keyword := range programmingKeywords {
			if strings.Contains(ua, keyword) {
				return true
			}
		}
	}

	return false
}

// 检测登录状态
func (c *Collector) detectLoginStatus(r *http.Request) bool {
	// 检查常见的登录相关cookie
	loginCookies := []string{"session", "token", "auth", "login", "user_id", "jwt"}
	
	for _, cookieName := range loginCookies {
		if cookie, err := r.Cookie(cookieName); err == nil && cookie.Value != "" {
			return true
		}
	}

	// 检查Authorization头
	if auth := r.Header.Get("Authorization"); auth != "" {
		return true
	}

	return false
}

// 提取代理相关头信息
func (c *Collector) extractProxyHeaders(r *http.Request) map[string]string {
	proxyHeaders := make(map[string]string)
	
	// 代理相关的头列表
	proxyHeaderNames := []string{
		"X-Forwarded-For", "X-Real-IP", "X-Forwarded-Proto", "X-Forwarded-Host",
		"X-Forwarded-Port", "X-Forwarded-Server", "X-Client-IP", "CF-Connecting-IP",
		"True-Client-IP", "X-Cluster-Client-IP", "Forwarded", "Via",
		"X-Originating-IP", "X-Remote-IP", "X-Remote-Addr",
	}

	for _, headerName := range proxyHeaderNames {
		if value := r.Header.Get(headerName); value != "" {
			proxyHeaders[headerName] = value
		}
	}

	return proxyHeaders
}

// 检测是否通过代理
func (c *Collector) detectProxy(r *http.Request) bool {
	// 检查常见的代理头
	proxyIndicators := []string{
		"X-Forwarded-For", "X-Real-IP", "X-Forwarded-Proto",
		"CF-Connecting-IP", "True-Client-IP", "Via", "Forwarded",
	}

	for _, header := range proxyIndicators {
		if r.Header.Get(header) != "" {
			return true
		}
	}

	// 检查Via头（HTTP代理标准头）
	if via := r.Header.Get("Via"); via != "" {
		return true
	}

	// 检查Connection头中的代理标识
	if conn := r.Header.Get("Connection"); strings.Contains(strings.ToLower(conn), "proxy") {
		return true
	}

	return false
}

// 提取代理链信息
func (c *Collector) extractProxyChain(r *http.Request) []string {
	var proxyChain []string

	// 从X-Forwarded-For提取代理链
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				proxyChain = append(proxyChain, ip)
			}
		}
	}

	// 从Via头提取代理信息
	if via := r.Header.Get("Via"); via != "" {
		// Via: 1.1 proxy1.example.com, 1.0 proxy2.example.com
		vias := strings.Split(via, ",")
		for _, v := range vias {
			v = strings.TrimSpace(v)
			if v != "" {
				// 提取代理服务器信息
				parts := strings.Fields(v)
				if len(parts) >= 2 {
					proxyChain = append(proxyChain, parts[1]) // 代理服务器名
				}
			}
		}
	}

	// 从Forwarded头提取（RFC 7239）
	if forwarded := r.Header.Get("Forwarded"); forwarded != "" {
		// Forwarded: for=192.0.2.60;proto=http;by=203.0.113.43
		parts := strings.Split(forwarded, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			// 解析by参数（代理服务器）
			if strings.Contains(part, "by=") {
				byStart := strings.Index(part, "by=") + 3
				byEnd := strings.Index(part[byStart:], ";")
				if byEnd == -1 {
					byEnd = len(part[byStart:])
				}
				proxyServer := strings.Trim(part[byStart:byStart+byEnd], "\"'")
				if proxyServer != "" {
					proxyChain = append(proxyChain, proxyServer)
				}
			}
		}
	}

	return proxyChain
}

// 获取代理信息摘要
func (c *Collector) GetProxySummary(info *AccessInfo) map[string]interface{} {
	return map[string]interface{}{
		"is_behind_proxy": info.IsBehindProxy,
		"original_ip":     info.OriginalIP,
		"real_ip":         info.IP,
		"proxy_chain":     info.ProxyChain,
		"proxy_count":     len(info.ProxyChain),
		"proxy_headers":   info.ProxyHeaders,
	}
}

// 获取访问统计信息
func (c *Collector) GetAccessSummary(info *AccessInfo) map[string]interface{} {
	return map[string]interface{}{
		"ip":             info.IP,
		"original_ip":    info.OriginalIP,
		"device_type":    info.DeviceType,
		"network_type":   info.NetworkType,
		"is_bot":         info.IsBot,
		"login_status":   info.LoginStatus,
		"is_behind_proxy": info.IsBehindProxy,
		"proxy_count":    len(info.ProxyChain),
		"path":           info.Path,
		"method":         info.Method,
		"timestamp":      info.Timestamp.Unix(),
	}
}
