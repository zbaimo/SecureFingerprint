package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLClient struct {
	db *sql.DB
}

type AccessRecord struct {
	ID          int       `json:"id"`
	Fingerprint string    `json:"fingerprint"`
	IP          string    `json:"ip"`
	UserAgent   string    `json:"user_agent"`
	Path        string    `json:"path"`
	Method      string    `json:"method"`
	Score       int       `json:"score"`
	Action      string    `json:"action"` // "allow", "limit", "ban"
	Timestamp   time.Time `json:"timestamp"`
}

type UserStats struct {
	Fingerprint    string    `json:"fingerprint"`
	TotalRequests  int       `json:"total_requests"`
	CurrentScore   int       `json:"current_score"`
	FirstSeen      time.Time `json:"first_seen"`
	LastSeen       time.Time `json:"last_seen"`
	BanCount       int       `json:"ban_count"`
}

func NewMySQLClient(dsn string, maxOpenConns, maxIdleConns int, connMaxLifetime time.Duration) (*MySQLClient, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("MySQL连接失败: %v", err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("MySQL ping失败: %v", err)
	}

	client := &MySQLClient{db: db}
	if err := client.initTables(); err != nil {
		return nil, fmt.Errorf("初始化数据表失败: %v", err)
	}

	return client, nil
}

func (m *MySQLClient) initTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS access_logs (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			fingerprint VARCHAR(64) NOT NULL,
			ip VARCHAR(45) NOT NULL,
			user_agent TEXT,
			path VARCHAR(500),
			method VARCHAR(10),
			score INT DEFAULT 100,
			action VARCHAR(20) DEFAULT 'allow',
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_fingerprint (fingerprint),
			INDEX idx_timestamp (timestamp),
			INDEX idx_ip (ip)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE IF NOT EXISTS user_stats (
			fingerprint VARCHAR(64) PRIMARY KEY,
			total_requests INT DEFAULT 0,
			current_score INT DEFAULT 100,
			first_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			ban_count INT DEFAULT 0
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE IF NOT EXISTS ban_history (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			fingerprint VARCHAR(64) NOT NULL,
			reason VARCHAR(200),
			banned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			unbanned_at TIMESTAMP NULL,
			duration_seconds INT,
			INDEX idx_fingerprint (fingerprint),
			INDEX idx_banned_at (banned_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}

	for _, query := range queries {
		if _, err := m.db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

// 记录访问日志
func (m *MySQLClient) LogAccess(record *AccessRecord) error {
	query := `INSERT INTO access_logs (fingerprint, ip, user_agent, path, method, score, action, timestamp) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err := m.db.Exec(query, record.Fingerprint, record.IP, record.UserAgent, 
		record.Path, record.Method, record.Score, record.Action, record.Timestamp)
	
	if err != nil {
		return err
	}

	// 更新用户统计
	return m.updateUserStats(record.Fingerprint, record.Score)
}

// 更新用户统计信息
func (m *MySQLClient) updateUserStats(fingerprint string, score int) error {
	query := `INSERT INTO user_stats (fingerprint, total_requests, current_score, first_seen, last_seen) 
			  VALUES (?, 1, ?, NOW(), NOW()) 
			  ON DUPLICATE KEY UPDATE 
			  total_requests = total_requests + 1,
			  current_score = ?,
			  last_seen = NOW()`
	
	_, err := m.db.Exec(query, fingerprint, score, score)
	return err
}

// 获取用户统计信息
func (m *MySQLClient) GetUserStats(fingerprint string) (*UserStats, error) {
	query := `SELECT fingerprint, total_requests, current_score, first_seen, last_seen, ban_count 
			  FROM user_stats WHERE fingerprint = ?`
	
	var stats UserStats
	err := m.db.QueryRow(query, fingerprint).Scan(
		&stats.Fingerprint, &stats.TotalRequests, &stats.CurrentScore,
		&stats.FirstSeen, &stats.LastSeen, &stats.BanCount,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	return &stats, err
}

// 获取访问日志（支持分页和筛选）
func (m *MySQLClient) GetAccessLogs(fingerprint string, limit, offset int, startTime, endTime time.Time) ([]AccessRecord, error) {
	query := `SELECT id, fingerprint, ip, user_agent, path, method, score, action, timestamp 
			  FROM access_logs WHERE 1=1`
	args := []interface{}{}

	if fingerprint != "" {
		query += " AND fingerprint = ?"
		args = append(args, fingerprint)
	}

	if !startTime.IsZero() {
		query += " AND timestamp >= ?"
		args = append(args, startTime)
	}

	if !endTime.IsZero() {
		query += " AND timestamp <= ?"
		args = append(args, endTime)
	}

	query += " ORDER BY timestamp DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []AccessRecord
	for rows.Next() {
		var record AccessRecord
		err := rows.Scan(&record.ID, &record.Fingerprint, &record.IP, 
			&record.UserAgent, &record.Path, &record.Method, 
			&record.Score, &record.Action, &record.Timestamp)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// 记录封禁历史
func (m *MySQLClient) LogBan(fingerprint, reason string, durationSeconds int) error {
	query := `INSERT INTO ban_history (fingerprint, reason, banned_at, duration_seconds) 
			  VALUES (?, ?, NOW(), ?)`
	
	_, err := m.db.Exec(query, fingerprint, reason, durationSeconds)
	if err != nil {
		return err
	}

	// 更新用户统计中的封禁次数
	updateQuery := `UPDATE user_stats SET ban_count = ban_count + 1 WHERE fingerprint = ?`
	_, err = m.db.Exec(updateQuery, fingerprint)
	return err
}

// 获取系统统计信息
func (m *MySQLClient) GetSystemStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总访问次数
	var totalAccess int
	err := m.db.QueryRow("SELECT COUNT(*) FROM access_logs").Scan(&totalAccess)
	if err != nil {
		return nil, err
	}
	stats["total_access"] = totalAccess

	// 今日访问次数
	var todayAccess int
	err = m.db.QueryRow("SELECT COUNT(*) FROM access_logs WHERE DATE(timestamp) = CURDATE()").Scan(&todayAccess)
	if err != nil {
		return nil, err
	}
	stats["today_access"] = todayAccess

	// 总用户数
	var totalUsers int
	err = m.db.QueryRow("SELECT COUNT(*) FROM user_stats").Scan(&totalUsers)
	if err != nil {
		return nil, err
	}
	stats["total_users"] = totalUsers

	// 被封禁用户数
	var bannedUsers int
	err = m.db.QueryRow("SELECT COUNT(DISTINCT fingerprint) FROM ban_history WHERE unbanned_at IS NULL").Scan(&bannedUsers)
	if err != nil {
		return nil, err
	}
	stats["banned_users"] = bannedUsers

	return stats, nil
}

func (m *MySQLClient) Close() error {
	return m.db.Close()
}
