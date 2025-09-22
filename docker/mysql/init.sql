-- 初始化防火墙控制器数据库

USE firewall_controller;

-- 创建访问日志表
CREATE TABLE IF NOT EXISTS access_logs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    fingerprint VARCHAR(64) NOT NULL COMMENT '用户指纹',
    ip VARCHAR(45) NOT NULL COMMENT 'IP地址',
    user_agent TEXT COMMENT 'User-Agent',
    path VARCHAR(500) COMMENT '访问路径',
    method VARCHAR(10) COMMENT 'HTTP方法',
    score INT DEFAULT 100 COMMENT '用户分数',
    action VARCHAR(20) DEFAULT 'allow' COMMENT '处理动作',
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '访问时间',
    INDEX idx_fingerprint (fingerprint),
    INDEX idx_timestamp (timestamp),
    INDEX idx_ip (ip),
    INDEX idx_action (action),
    INDEX idx_score (score)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='访问日志表';

-- 创建用户统计表
CREATE TABLE IF NOT EXISTS user_stats (
    fingerprint VARCHAR(64) PRIMARY KEY COMMENT '用户指纹',
    total_requests INT DEFAULT 0 COMMENT '总请求数',
    current_score INT DEFAULT 100 COMMENT '当前分数',
    first_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '首次访问时间',
    last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后访问时间',
    ban_count INT DEFAULT 0 COMMENT '封禁次数',
    INDEX idx_score (current_score),
    INDEX idx_last_seen (last_seen),
    INDEX idx_ban_count (ban_count)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户统计表';

-- 创建封禁历史表
CREATE TABLE IF NOT EXISTS ban_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    fingerprint VARCHAR(64) NOT NULL COMMENT '用户指纹',
    ip VARCHAR(45) COMMENT 'IP地址',
    reason VARCHAR(200) COMMENT '封禁原因',
    banned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '封禁时间',
    unbanned_at TIMESTAMP NULL COMMENT '解封时间',
    duration_seconds INT COMMENT '封禁时长(秒)',
    operator VARCHAR(50) DEFAULT 'system' COMMENT '操作员',
    INDEX idx_fingerprint (fingerprint),
    INDEX idx_banned_at (banned_at),
    INDEX idx_ip (ip)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='封禁历史表';

-- 创建配置历史表
CREATE TABLE IF NOT EXISTS config_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    config_type VARCHAR(50) NOT NULL COMMENT '配置类型',
    old_value TEXT COMMENT '旧配置值',
    new_value TEXT COMMENT '新配置值',
    changes TEXT COMMENT '变更描述',
    operator VARCHAR(50) DEFAULT 'system' COMMENT '操作员',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_config_type (config_type),
    INDEX idx_created_at (created_at),
    INDEX idx_operator (operator)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='配置历史表';

-- 创建系统事件表
CREATE TABLE IF NOT EXISTS system_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL COMMENT '事件类型',
    event_level VARCHAR(20) DEFAULT 'info' COMMENT '事件级别',
    title VARCHAR(200) NOT NULL COMMENT '事件标题',
    description TEXT COMMENT '事件描述',
    data JSON COMMENT '事件数据',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_event_type (event_type),
    INDEX idx_event_level (event_level),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统事件表';

-- 插入初始数据
INSERT INTO system_events (event_type, event_level, title, description) VALUES
('system', 'info', '系统初始化', '防火墙控制器系统初始化完成'),
('config', 'info', '默认配置加载', '系统默认配置已加载');

-- 创建视图：用户风险视图
CREATE OR REPLACE VIEW user_risk_view AS
SELECT 
    us.fingerprint,
    us.current_score,
    us.total_requests,
    us.ban_count,
    us.last_seen,
    CASE 
        WHEN us.current_score >= 80 THEN 'low'
        WHEN us.current_score >= 60 THEN 'medium'
        WHEN us.current_score >= 30 THEN 'high'
        ELSE 'critical'
    END as risk_level,
    (SELECT COUNT(*) FROM access_logs al WHERE al.fingerprint = us.fingerprint AND al.action = 'ban' AND al.timestamp >= DATE_SUB(NOW(), INTERVAL 24 HOUR)) as recent_bans
FROM user_stats us;

-- 创建视图：访问统计视图
CREATE OR REPLACE VIEW access_stats_view AS
SELECT 
    DATE(timestamp) as date,
    COUNT(*) as total_requests,
    COUNT(CASE WHEN action = 'allow' THEN 1 END) as allowed_requests,
    COUNT(CASE WHEN action = 'limit' THEN 1 END) as limited_requests,
    COUNT(CASE WHEN action = 'challenge' THEN 1 END) as challenged_requests,
    COUNT(CASE WHEN action = 'ban' THEN 1 END) as banned_requests,
    COUNT(DISTINCT fingerprint) as unique_users,
    COUNT(DISTINCT ip) as unique_ips
FROM access_logs 
WHERE timestamp >= DATE_SUB(NOW(), INTERVAL 30 DAY)
GROUP BY DATE(timestamp)
ORDER BY date DESC;

-- 创建存储过程：清理过期数据
DELIMITER //
CREATE PROCEDURE CleanupExpiredData(IN days_to_keep INT)
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    START TRANSACTION;
    
    -- 清理过期的访问日志
    DELETE FROM access_logs WHERE timestamp < DATE_SUB(NOW(), INTERVAL days_to_keep DAY);
    
    -- 清理过期的封禁历史
    DELETE FROM ban_history WHERE banned_at < DATE_SUB(NOW(), INTERVAL days_to_keep DAY);
    
    -- 清理过期的配置历史
    DELETE FROM config_history WHERE created_at < DATE_SUB(NOW(), INTERVAL days_to_keep DAY);
    
    -- 清理过期的系统事件
    DELETE FROM system_events WHERE created_at < DATE_SUB(NOW(), INTERVAL days_to_keep DAY);
    
    COMMIT;
    
    SELECT 'Data cleanup completed' as result;
END //
DELIMITER ;

-- 创建定时任务清理过期数据（需要开启事件调度器）
-- SET GLOBAL event_scheduler = ON;
-- CREATE EVENT IF NOT EXISTS cleanup_expired_data
-- ON SCHEDULE EVERY 1 DAY
-- STARTS CURRENT_TIMESTAMP
-- DO CALL CleanupExpiredData(30);

COMMIT;
