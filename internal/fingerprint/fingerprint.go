package fingerprint

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"sort"
	"strings"

	"securefingerprint/internal/collector"
)

type Generator struct {
	salt string // 用于增加指纹安全性的盐值
}

// 指纹组件权重配置
type FingerprintWeights struct {
	IP        float64 `json:"ip"`         // IP地址权重
	UserAgent float64 `json:"user_agent"` // User-Agent权重
	Headers   float64 `json:"headers"`    // HTTP头权重
	Network   float64 `json:"network"`    // 网络类型权重
	Device    float64 `json:"device"`     // 设备类型权重
}

// 默认权重配置
var DefaultWeights = FingerprintWeights{
	IP:        0.4,  // IP地址占40%
	UserAgent: 0.3,  // User-Agent占30%
	Headers:   0.15, // HTTP头占15%
	Network:   0.1,  // 网络类型占10%
	Device:    0.05, // 设备类型占5%
}

func NewGenerator(salt string) *Generator {
	if salt == "" {
		salt = "firewall-controller-default-salt"
	}
	return &Generator{salt: salt}
}

// 生成用户指纹
func (g *Generator) Generate(info *collector.AccessInfo) string {
	return g.GenerateWithWeights(info, DefaultWeights)
}

// 使用自定义权重生成指纹
func (g *Generator) GenerateWithWeights(info *collector.AccessInfo, weights FingerprintWeights) string {
	components := g.extractComponents(info)
	
	// 根据权重组合指纹组件
	fingerprintData := g.combineComponents(components, weights)
	
	// 生成最终指纹
	return g.hashFingerprint(fingerprintData)
}

// 提取指纹组件
func (g *Generator) extractComponents(info *collector.AccessInfo) map[string]string {
	components := make(map[string]string)

	// IP地址组件（对IP进行部分模糊化以增加稳定性）
	components["ip"] = g.normalizeIP(info.IP)
	
	// User-Agent组件（提取关键特征）
	components["user_agent"] = g.normalizeUserAgent(info.UserAgent)
	
	// HTTP头组件（选择稳定的头信息）
	components["headers"] = g.normalizeHeaders(info.Headers)
	
	// 网络类型组件
	components["network"] = info.NetworkType
	
	// 设备类型组件
	components["device"] = info.DeviceType

	return components
}

// 规范化IP地址（保留网段信息，增加稳定性）
func (g *Generator) normalizeIP(ip string) string {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return ip
	}

	// 对于IPv4，保留前3个八位组
	if ipv4 := parsedIP.To4(); ipv4 != nil {
		return fmt.Sprintf("%d.%d.%d.0", ipv4[0], ipv4[1], ipv4[2])
	}

	// 对于IPv6，保留前64位
	if ipv6 := parsedIP.To16(); ipv6 != nil {
		return fmt.Sprintf("%02x%02x:%02x%02x:%02x%02x:%02x%02x::",
			ipv6[0], ipv6[1], ipv6[2], ipv6[3],
			ipv6[4], ipv6[5], ipv6[6], ipv6[7])
	}

	return ip
}

// 规范化User-Agent（提取关键特征，忽略版本号细节）
func (g *Generator) normalizeUserAgent(userAgent string) string {
	if userAgent == "" {
		return "empty"
	}

	ua := strings.ToLower(userAgent)
	
	// 提取浏览器主要信息
	var browser, os, engine string
	
	// 检测浏览器
	if strings.Contains(ua, "chrome") {
		browser = "chrome"
	} else if strings.Contains(ua, "firefox") {
		browser = "firefox"
	} else if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		browser = "safari"
	} else if strings.Contains(ua, "edge") {
		browser = "edge"
	} else if strings.Contains(ua, "opera") {
		browser = "opera"
	} else {
		browser = "other"
	}
	
	// 检测操作系统
	if strings.Contains(ua, "windows") {
		os = "windows"
	} else if strings.Contains(ua, "mac os") || strings.Contains(ua, "macos") {
		os = "macos"
	} else if strings.Contains(ua, "linux") {
		os = "linux"
	} else if strings.Contains(ua, "android") {
		os = "android"
	} else if strings.Contains(ua, "ios") || strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") {
		os = "ios"
	} else {
		os = "other"
	}
	
	// 检测渲染引擎
	if strings.Contains(ua, "webkit") {
		engine = "webkit"
	} else if strings.Contains(ua, "gecko") {
		engine = "gecko"
	} else if strings.Contains(ua, "trident") {
		engine = "trident"
	} else {
		engine = "other"
	}
	
	return fmt.Sprintf("%s_%s_%s", browser, os, engine)
}

// 规范化HTTP头信息
func (g *Generator) normalizeHeaders(headers map[string]string) string {
	if len(headers) == 0 {
		return "empty"
	}

	// 选择稳定的头信息进行指纹生成
	stableHeaders := []string{
		"Accept", "Accept-Language", "Accept-Encoding",
		"DNT", "Upgrade-Insecure-Requests",
	}

	var headerParts []string
	for _, header := range stableHeaders {
		if value, exists := headers[header]; exists {
			// 简化Accept-Language（只保留主要语言）
			if header == "Accept-Language" {
				value = g.normalizeAcceptLanguage(value)
			}
			headerParts = append(headerParts, fmt.Sprintf("%s:%s", header, value))
		}
	}

	// 排序确保一致性
	sort.Strings(headerParts)
	
	if len(headerParts) == 0 {
		return "empty"
	}
	
	return strings.Join(headerParts, "|")
}

// 规范化Accept-Language头
func (g *Generator) normalizeAcceptLanguage(acceptLang string) string {
	if acceptLang == "" {
		return "empty"
	}

	// 提取主要语言（忽略权重和地区变体）
	parts := strings.Split(acceptLang, ",")
	var langs []string
	
	for _, part := range parts {
		lang := strings.TrimSpace(part)
		// 移除权重信息
		if idx := strings.Index(lang, ";"); idx != -1 {
			lang = lang[:idx]
		}
		// 只保留语言代码，忽略地区代码
		if idx := strings.Index(lang, "-"); idx != -1 {
			lang = lang[:idx]
		}
		if lang != "" {
			langs = append(langs, lang)
		}
	}

	// 去重并排序
	langMap := make(map[string]bool)
	for _, lang := range langs {
		langMap[lang] = true
	}
	
	var uniqueLangs []string
	for lang := range langMap {
		uniqueLangs = append(uniqueLangs, lang)
	}
	sort.Strings(uniqueLangs)

	if len(uniqueLangs) == 0 {
		return "empty"
	}
	
	return strings.Join(uniqueLangs, ",")
}

// 根据权重组合指纹组件
func (g *Generator) combineComponents(components map[string]string, weights FingerprintWeights) string {
	var parts []string
	
	// 按权重添加组件
	if weights.IP > 0 {
		parts = append(parts, fmt.Sprintf("ip:%.2f:%s", weights.IP, components["ip"]))
	}
	if weights.UserAgent > 0 {
		parts = append(parts, fmt.Sprintf("ua:%.2f:%s", weights.UserAgent, components["user_agent"]))
	}
	if weights.Headers > 0 {
		parts = append(parts, fmt.Sprintf("hdr:%.2f:%s", weights.Headers, components["headers"]))
	}
	if weights.Network > 0 {
		parts = append(parts, fmt.Sprintf("net:%.2f:%s", weights.Network, components["network"]))
	}
	if weights.Device > 0 {
		parts = append(parts, fmt.Sprintf("dev:%.2f:%s", weights.Device, components["device"]))
	}

	return strings.Join(parts, "|")
}

// 生成最终哈希指纹
func (g *Generator) hashFingerprint(data string) string {
	// 添加盐值
	saltedData := fmt.Sprintf("%s|salt:%s", data, g.salt)
	
	// 使用SHA256生成哈希
	hasher := sha256.New()
	hasher.Write([]byte(saltedData))
	hash := hasher.Sum(nil)
	
	// 转换为16进制字符串
	return hex.EncodeToString(hash)
}

// 生成短指纹（用于显示）
func (g *Generator) GenerateShort(info *collector.AccessInfo) string {
	fullFingerprint := g.Generate(info)
	
	// 使用MD5生成较短的指纹
	hasher := md5.New()
	hasher.Write([]byte(fullFingerprint))
	hash := hasher.Sum(nil)
	
	// 返回前16个字符
	return hex.EncodeToString(hash)[:16]
}

// 验证指纹格式
func (g *Generator) ValidateFingerprint(fingerprint string) bool {
	// 检查长度（SHA256的十六进制表示应该是64个字符）
	if len(fingerprint) != 64 {
		return false
	}
	
	// 检查是否为有效的十六进制字符串
	_, err := hex.DecodeString(fingerprint)
	return err == nil
}

// 比较两个指纹的相似度
func (g *Generator) CalculateSimilarity(fp1, fp2 string) float64 {
	if !g.ValidateFingerprint(fp1) || !g.ValidateFingerprint(fp2) {
		return 0.0
	}
	
	if fp1 == fp2 {
		return 1.0
	}
	
	// 计算汉明距离
	bytes1, _ := hex.DecodeString(fp1)
	bytes2, _ := hex.DecodeString(fp2)
	
	if len(bytes1) != len(bytes2) {
		return 0.0
	}
	
	var differentBits int
	for i := 0; i < len(bytes1); i++ {
		xor := bytes1[i] ^ bytes2[i]
		// 计算字节中的1的个数
		for xor != 0 {
			differentBits++
			xor &= xor - 1
		}
	}
	
	totalBits := len(bytes1) * 8
	similarity := 1.0 - float64(differentBits)/float64(totalBits)
	
	return similarity
}

// 获取指纹详细信息（用于调试）
func (g *Generator) GetFingerprintDetails(info *collector.AccessInfo) map[string]interface{} {
	components := g.extractComponents(info)
	
	return map[string]interface{}{
		"components": components,
		"weights":    DefaultWeights,
		"combined":   g.combineComponents(components, DefaultWeights),
		"fingerprint": g.Generate(info),
		"short_fingerprint": g.GenerateShort(info),
	}
}
