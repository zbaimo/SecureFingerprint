# Nginx代理环境配置指南

本文档介绍如何在nginx代理环境下正确配置防火墙控制器以获取真实客户端信息。

## 🔧 Nginx配置

### 1. 基本代理配置

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        # 传递真实客户端IP
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-Port $server_port;
        
        # 传递原始Host头
        proxy_set_header Host $host;
        
        # 支持WebSocket
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # 代理到防火墙控制器
        proxy_pass http://firewall-controller:8080;
        
        # 超时设置
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }
}
```

### 2. 多层代理配置

如果您有多层代理（如CDN + Nginx），需要特殊处理：

```nginx
# 在http块中定义可信代理
http {
    # 设置可信代理IP范围
    set_real_ip_from 10.0.0.0/8;
    set_real_ip_from 172.16.0.0/12;
    set_real_ip_from 192.168.0.0/16;
    set_real_ip_from 127.0.0.1;
    
    # Cloudflare IP范围（示例）
    set_real_ip_from 173.245.48.0/20;
    set_real_ip_from 103.21.244.0/22;
    set_real_ip_from 103.22.200.0/22;
    # ... 更多Cloudflare IP范围
    
    # 指定真实IP头
    real_ip_header X-Forwarded-For;
    real_ip_recursive on;
    
    server {
        # ... 其他配置
        
        location / {
            # 传递处理后的真实IP
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Original-Forwarded-For $http_x_forwarded_for;
            
            # 传递代理链信息
            proxy_set_header X-Proxy-Chain $proxy_add_x_forwarded_for;
            
            proxy_pass http://firewall-controller:8080;
        }
    }
}
```

### 3. Cloudflare配置

如果使用Cloudflare，需要特殊处理：

```nginx
server {
    location / {
        # Cloudflare提供的真实IP头
        proxy_set_header X-Real-IP $http_cf_connecting_ip;
        proxy_set_header CF-Connecting-IP $http_cf_connecting_ip;
        
        # 标准代理头
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Cloudflare特有头
        proxy_set_header CF-Ray $http_cf_ray;
        proxy_set_header CF-Visitor $http_cf_visitor;
        proxy_set_header CF-Country $http_cf_ipcountry;
        
        proxy_pass http://firewall-controller:8080;
    }
}
```

## ⚙️ 防火墙控制器配置

### 1. 基本配置

在 `configs/config.yaml` 中配置代理设置：

```yaml
security:
  proxy:
    trusted_proxies:
      - "127.0.0.1/32"      # 本地
      - "10.0.0.0/8"        # 内网A类
      - "172.16.0.0/12"     # 内网B类
      - "192.168.0.0/16"    # 内网C类
      - "::1/128"           # IPv6回环
    
    trusted_headers:
      - "X-Real-IP"
      - "X-Forwarded-For"
      - "CF-Connecting-IP"
      - "True-Client-IP"
    
    header_priority:
      CF-Connecting-IP: 100   # Cloudflare最高优先级
      True-Client-IP: 90      # Akamai
      X-Real-IP: 80          # Nginx
      X-Forwarded-For: 70    # 标准代理头
    
    skip_private_ranges: true
    max_proxy_depth: 10
```

### 2. Cloudflare配置

如果使用Cloudflare CDN：

```yaml
security:
  proxy:
    trusted_proxies:
      # Cloudflare IP范围
      - "173.245.48.0/20"
      - "103.21.244.0/22"
      - "103.22.200.0/22"
      - "103.31.4.0/22"
      - "141.101.64.0/18"
      - "108.162.192.0/18"
      - "190.93.240.0/20"
      - "188.114.96.0/20"
      - "197.234.240.0/22"
      - "198.41.128.0/17"
      - "162.158.0.0/15"
      - "104.16.0.0/13"
      - "104.24.0.0/14"
      - "172.64.0.0/13"
      - "131.0.72.0/22"
    
    trusted_headers:
      - "CF-Connecting-IP"    # Cloudflare真实IP
      - "X-Forwarded-For"
      - "X-Real-IP"
    
    header_priority:
      CF-Connecting-IP: 100
      X-Forwarded-For: 80
      X-Real-IP: 70
```

## 🧪 测试配置

### 1. 验证IP获取

创建测试端点来验证IP获取是否正确：

```bash
# 直接访问（应该显示您的真实IP）
curl -H "Host: your-domain.com" http://your-server/api/v1/system/info

# 通过代理访问
curl -H "X-Forwarded-For: 1.2.3.4, 192.168.1.1" \
     -H "X-Real-IP: 1.2.3.4" \
     http://your-server/api/v1/system/info
```

### 2. 检查代理头

使用以下命令检查代理头是否正确传递：

```bash
# 查看访问日志中的IP信息
curl http://your-domain.com/api/v1/logs/recent?limit=1

# 查看详细的代理信息
curl http://your-domain.com/api/v1/logs/user/YOUR_FINGERPRINT
```

### 3. 验证代理检测

```javascript
// 在浏览器控制台中测试
fetch('/api/v1/system/info')
  .then(r => r.json())
  .then(data => {
    console.log('系统信息:', data);
    // 检查是否正确识别了代理环境
  });
```

## 🔍 故障排除

### 1. IP显示为内网地址

**问题**: 获取到的IP是 `127.0.0.1` 或 `192.168.x.x`

**解决方案**:
- 检查nginx是否正确设置了 `proxy_set_header X-Real-IP $remote_addr`
- 确认防火墙控制器的可信代理配置包含了nginx的IP
- 验证 `real_ip_header` 配置是否正确

### 2. 代理检测不生效

**问题**: 系统显示 `is_behind_proxy: false`

**解决方案**:
- 检查代理头是否正确传递到应用
- 确认 `trusted_headers` 配置包含了nginx使用的头
- 验证nginx配置中的头名称是否正确

### 3. 多层代理问题

**问题**: 在CDN+Nginx环境下获取不到真实IP

**解决方案**:
- 配置nginx的 `real_ip_header` 和 `set_real_ip_from`
- 在防火墙控制器中配置CDN的IP范围为可信代理
- 使用CDN特有的头（如 `CF-Connecting-IP`）

### 4. IPv6支持

**问题**: IPv6地址处理异常

**解决方案**:
- 确保nginx配置支持IPv6: `listen [::]:80`
- 在可信代理中添加IPv6范围: `::1/128`, `fc00::/7`
- 检查IP解析逻辑是否支持IPv6格式

## 📝 最佳实践

1. **安全性**: 只信任已知的代理服务器IP
2. **性能**: 限制代理链深度避免过度处理
3. **监控**: 定期检查代理配置是否正常工作
4. **日志**: 记录代理相关信息便于调试
5. **更新**: 及时更新CDN提供商的IP范围

## 🔗 相关链接

- [Nginx real_ip 模块文档](http://nginx.org/en/docs/http/ngx_http_realip_module.html)
- [RFC 7239 - Forwarded HTTP Extension](https://tools.ietf.org/html/rfc7239)
- [Cloudflare IP 范围](https://www.cloudflare.com/ips/)
- [X-Forwarded-For 标准](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For)
