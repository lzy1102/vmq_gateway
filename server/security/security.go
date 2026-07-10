package security

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// ValidateCallbackURL 验证回调地址，防止 SSRF
func ValidateCallbackURL(rawURL string) error {
	if rawURL == "" {
		return nil // 可选字段
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("无效的 URL")
	}

	// 只允许 http/https
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("仅支持 http/https 协议")
	}

	hostname := u.Hostname()
	if hostname == "" {
		return fmt.Errorf("无效的主机名")
	}

	// 检查是否为内网地址
	if isPrivateIP(hostname) {
		return fmt.Errorf("禁止访问内网地址")
	}

	// 禁止 localhost 变体
	lower := strings.ToLower(hostname)
	if lower == "localhost" || lower == "0.0.0.0" || lower == "127.0.0.1" || lower == "::1" {
		return fmt.Errorf("禁止访问本地地址")
	}

	// 禁止 metadata 地址
	if lower == "169.254.169.254" || lower == "metadata.google.internal" {
		return fmt.Errorf("禁止访问元数据服务")
	}

	return nil
}

// isPrivateIP 判断是否为内网/私有 IP
func isPrivateIP(hostname string) bool {
	ip := net.ParseIP(hostname)
	if ip == nil {
		// 可能是域名，不拦截
		return false
	}

	// 回环地址
	if ip.IsLoopback() {
		return true
	}

	// 私有地址范围
	if ip.IsPrivate() {
		return true
	}

	// 链路本地
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	// 0.0.0.0
	if ip.IsUnspecified() {
		return true
	}

	return false
}
