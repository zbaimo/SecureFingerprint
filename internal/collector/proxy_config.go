package collector

import (
	"net"
	"net/http"
	"strings"
)

// 代理配置
type ProxyConfig struct {
	TrustedProxies    []string          `yaml:"trusted_proxies"`    // 可信代理IP/CIDR列表
	TrustedHeaders    []string          `yaml:"trusted_headers"`    // 可信的代理头列表
	HeaderPriority    map[string]int    `yaml:"header_priority"`    // 头优先级配置
	SkipPrivateRanges bool              `yaml:"skip_private_ranges"` // 是否跳过内网IP
	MaxProxyDepth     int               `yaml:"max_proxy_depth"`     // 最大代理深度
}

// 默认代理配置
var DefaultProxyConfig = ProxyConfig{
	TrustedProxies: []string{
		"127.0.0.1/32",      // 本地回环
		"10.0.0.0/8",        // 内网A类
		"172.16.0.0/12",     // 内网B类
		"192.168.0.0/16",    // 内网C类
		"::1/128",           // IPv6回环
		"fc00::/7",          // IPv6内网
	},
	TrustedHeaders: []string{
		"X-Real-IP",
		"X-Forwarded-For",
		"CF-Connecting-IP",
		"True-Client-IP",
	},
	HeaderPriority: map[string]int{
		"CF-Connecting-IP": 100, // Cloudflare最高优先级
		"True-Client-IP":   90,  // Akamai
		"X-Real-IP":        80,  // Nginx
		"X-Forwarded-For":  70,  // 标准代理头
		"X-Client-IP":      60,  // Apache
	},
	SkipPrivateRanges: true,
	MaxProxyDepth:     10,
}

// 代理检测器
type ProxyDetector struct {
	config       ProxyConfig
	trustedNets  []*net.IPNet
	headersByPriority []string
}

// 创建代理检测器
func NewProxyDetector(config ProxyConfig) (*ProxyDetector, error) {
	detector := &ProxyDetector{
		config: config,
	}

	// 解析可信代理网络
	for _, cidr := range config.TrustedProxies {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			// 尝试解析为单个IP
			ip := net.ParseIP(cidr)
			if ip != nil {
				if ip.To4() != nil {
					_, network, _ = net.ParseCIDR(cidr + "/32")
				} else {
					_, network, _ = net.ParseCIDR(cidr + "/128")
				}
			} else {
				return nil, err
			}
		}
		detector.trustedNets = append(detector.trustedNets, network)
	}

	// 按优先级排序头列表
	detector.sortHeadersByPriority()

	return detector, nil
}

// 按优先级排序头
func (pd *ProxyDetector) sortHeadersByPriority() {
	headerPriority := make(map[string]int)
	
	// 使用配置的优先级
	for header, priority := range pd.config.HeaderPriority {
		headerPriority[header] = priority
	}
	
	// 为未配置优先级的头设置默认值
	for _, header := range pd.config.TrustedHeaders {
		if _, exists := headerPriority[header]; !exists {
			headerPriority[header] = 50 // 默认优先级
		}
	}
	
	// 排序
	headers := make([]string, 0, len(headerPriority))
	for header := range headerPriority {
		headers = append(headers, header)
	}
	
	// 按优先级降序排序
	for i := 0; i < len(headers)-1; i++ {
		for j := i + 1; j < len(headers); j++ {
			if headerPriority[headers[i]] < headerPriority[headers[j]] {
				headers[i], headers[j] = headers[j], headers[i]
			}
		}
	}
	
	pd.headersByPriority = headers
}

// 检查IP是否为可信代理
func (pd *ProxyDetector) IsTrustedProxy(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	for _, network := range pd.trustedNets {
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// 从请求中提取真实客户端IP
func (pd *ProxyDetector) ExtractRealIP(r *http.Request) (string, []string) {
	originalIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	
	var proxyChain []string
	var realIP string

	// 如果直连的IP不是可信代理，直接返回
	if !pd.IsTrustedProxy(originalIP) {
		return originalIP, nil
	}

	// 按优先级检查头
	for _, header := range pd.headersByPriority {
		value := r.Header.Get(header)
		if value == "" {
			continue
		}

		ips := pd.parseIPsFromHeader(value, header)
		if len(ips) == 0 {
			continue
		}

		// 构建代理链
		proxyChain = append(proxyChain, ips...)

		// 从右到左找第一个非可信代理的IP作为真实客户端IP
		for i := len(ips) - 1; i >= 0; i-- {
			ip := strings.TrimSpace(ips[i])
			if ip != "" && net.ParseIP(ip) != nil {
				if !pd.IsTrustedProxy(ip) {
					realIP = ip
					break
				}
			}
		}

		if realIP != "" {
			break
		}
	}

	// 如果没找到真实IP，使用第一个有效IP
	if realIP == "" && len(proxyChain) > 0 {
		for _, ip := range proxyChain {
			ip = strings.TrimSpace(ip)
			if ip != "" && net.ParseIP(ip) != nil {
				if pd.config.SkipPrivateRanges {
					parsedIP := net.ParseIP(ip)
					if !isPrivateIP(parsedIP) {
						realIP = ip
						break
					}
				} else {
					realIP = ip
					break
				}
			}
		}
	}

	// 如果还是没找到，使用原始IP
	if realIP == "" {
		realIP = originalIP
	}

	// 限制代理链深度
	if len(proxyChain) > pd.config.MaxProxyDepth {
		proxyChain = proxyChain[:pd.config.MaxProxyDepth]
	}

	return realIP, proxyChain
}

// 从头中解析IP列表
func (pd *ProxyDetector) parseIPsFromHeader(value, header string) []string {
	var ips []string

	switch header {
	case "X-Forwarded-For":
		// X-Forwarded-For: client, proxy1, proxy2
		parts := strings.Split(value, ",")
		for _, part := range parts {
			ip := strings.TrimSpace(part)
			if ip != "" {
				ips = append(ips, ip)
			}
		}

	case "Forwarded":
		// RFC 7239: Forwarded: for=192.0.2.60;proto=http;by=203.0.113.43
		parts := strings.Split(value, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.Contains(part, "for=") {
				forStart := strings.Index(part, "for=") + 4
				forEnd := strings.Index(part[forStart:], ";")
				if forEnd == -1 {
					forEnd = len(part[forStart:])
				}
				forValue := strings.Trim(part[forStart:forStart+forEnd], "\"'")
				if forValue != "" {
					ips = append(ips, forValue)
				}
			}
		}

	default:
		// 其他头通常只包含单个IP
		ip := strings.TrimSpace(value)
		if ip != "" {
			ips = append(ips, ip)
		}
	}

	return ips
}

// 检查是否为内网IP
func isPrivateIP(ip net.IP) bool {
	if ip == nil {
		return false
	}

	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"::1/128",
		"fc00::/7",
	}

	for _, cidr := range privateRanges {
		_, network, _ := net.ParseCIDR(cidr)
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// 获取代理检测报告
func (pd *ProxyDetector) GetProxyReport(r *http.Request) map[string]interface{} {
	originalIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	realIP, proxyChain := pd.ExtractRealIP(r)
	
	report := map[string]interface{}{
		"original_ip":     originalIP,
		"real_ip":         realIP,
		"proxy_chain":     proxyChain,
		"proxy_count":     len(proxyChain),
		"is_behind_proxy": len(proxyChain) > 0,
		"is_trusted_proxy": pd.IsTrustedProxy(originalIP),
		"proxy_headers":   make(map[string]string),
	}

	// 收集代理头信息
	proxyHeaders := report["proxy_headers"].(map[string]string)
	for _, header := range pd.headersByPriority {
		if value := r.Header.Get(header); value != "" {
			proxyHeaders[header] = value
		}
	}

	return report
}
