package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// 访问记录查询条件
type AccessRecordQuery struct {
	Fingerprint string    `json:"fingerprint,omitempty"`
	IP          string    `json:"ip,omitempty"`
	UserAgent   string    `json:"user_agent,omitempty"`
	Path        string    `json:"path,omitempty"`
	Method      string    `json:"method,omitempty"`
	Action      string    `json:"action,omitempty"`
	MinScore    *int      `json:"min_score,omitempty"`
	MaxScore    *int      `json:"max_score,omitempty"`
	StartTime   time.Time `json:"start_time,omitempty"`
	EndTime     time.Time `json:"end_time,omitempty"`
	Limit       int       `json:"limit"`
	Offset      int       `json:"offset"`
	OrderBy     string    `json:"order_by"`
	OrderDir    string    `json:"order_dir"`
}

// 访问记录查询结果
type AccessRecordResult struct {
	Records    []AccessRecord `json:"records"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// 访问统计信息
type AccessStats struct {
	TotalRequests     int64                  `json:"total_requests"`
	UniqueUsers       int64                  `json:"unique_users"`
	UniqueIPs         int64                  `json:"unique_ips"`
	ActionStats       map[string]int64       `json:"action_stats"`
	HourlyStats       []HourlyAccessStat     `json:"hourly_stats"`
	PathStats         []PathAccessStat       `json:"path_stats"`
	UserAgentStats    []UserAgentStat        `json:"user_agent_stats"`
	ScoreDistribution []ScoreDistributionStat `json:"score_distribution"`
}

type HourlyAccessStat struct {
	Hour  int   `json:"hour"`
	Count int64 `json:"count"`
}

type PathAccessStat struct {
	Path  string `json:"path"`
	Count int64  `json:"count"`
}

type UserAgentStat struct {
	UserAgent string `json:"user_agent"`
	Count     int64  `json:"count"`
}

type ScoreDistributionStat struct {
	ScoreRange string `json:"score_range"`
	Count      int64  `json:"count"`
}

// 扩展MySQL客户端功能
func (m *MySQLClient) QueryAccessRecords(query *AccessRecordQuery) (*AccessRecordResult, error) {
	// 构建WHERE条件
	whereClause, args := m.buildWhereClause(query)
	
	// 获取总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM access_logs WHERE %s", whereClause)
	var total int64
	err := m.db.QueryRow(countSQL, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("查询总数失败: %v", err)
	}

	// 构建排序
	orderClause := m.buildOrderClause(query)
	
	// 构建分页
	limitClause := fmt.Sprintf("LIMIT %d OFFSET %d", query.Limit, query.Offset)
	
	// 查询数据
	dataSQL := fmt.Sprintf(`
		SELECT id, fingerprint, ip, user_agent, path, method, score, action, timestamp 
		FROM access_logs 
		WHERE %s 
		%s 
		%s`, whereClause, orderClause, limitClause)
	
	rows, err := m.db.Query(dataSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("查询数据失败: %v", err)
	}
	defer rows.Close()

	var records []AccessRecord
	for rows.Next() {
		var record AccessRecord
		err := rows.Scan(
			&record.ID, &record.Fingerprint, &record.IP, &record.UserAgent,
			&record.Path, &record.Method, &record.Score, &record.Action, &record.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描记录失败: %v", err)
		}
		records = append(records, record)
	}

	// 计算分页信息
	page := (query.Offset / query.Limit) + 1
	totalPages := int((total + int64(query.Limit) - 1) / int64(query.Limit))

	return &AccessRecordResult{
		Records:    records,
		Total:      total,
		Page:       page,
		PageSize:   query.Limit,
		TotalPages: totalPages,
	}, nil
}

// 构建WHERE子句
func (m *MySQLClient) buildWhereClause(query *AccessRecordQuery) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	// 基础条件
	conditions = append(conditions, "1=1")

	if query.Fingerprint != "" {
		conditions = append(conditions, "fingerprint = ?")
		args = append(args, query.Fingerprint)
	}

	if query.IP != "" {
		conditions = append(conditions, "ip = ?")
		args = append(args, query.IP)
	}

	if query.UserAgent != "" {
		conditions = append(conditions, "user_agent LIKE ?")
		args = append(args, "%"+query.UserAgent+"%")
	}

	if query.Path != "" {
		conditions = append(conditions, "path LIKE ?")
		args = append(args, "%"+query.Path+"%")
	}

	if query.Method != "" {
		conditions = append(conditions, "method = ?")
		args = append(args, query.Method)
	}

	if query.Action != "" {
		conditions = append(conditions, "action = ?")
		args = append(args, query.Action)
	}

	if query.MinScore != nil {
		conditions = append(conditions, "score >= ?")
		args = append(args, *query.MinScore)
	}

	if query.MaxScore != nil {
		conditions = append(conditions, "score <= ?")
		args = append(args, *query.MaxScore)
	}

	if !query.StartTime.IsZero() {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, query.StartTime)
	}

	if !query.EndTime.IsZero() {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, query.EndTime)
	}

	return strings.Join(conditions, " AND "), args
}

// 构建ORDER BY子句
func (m *MySQLClient) buildOrderClause(query *AccessRecordQuery) string {
	orderBy := "timestamp"
	if query.OrderBy != "" {
		// 验证排序字段，防止SQL注入
		validFields := map[string]bool{
			"id": true, "fingerprint": true, "ip": true, "score": true,
			"timestamp": true, "action": true, "method": true,
		}
		if validFields[query.OrderBy] {
			orderBy = query.OrderBy
		}
	}

	orderDir := "DESC"
	if query.OrderDir == "ASC" || query.OrderDir == "asc" {
		orderDir = "ASC"
	}

	return fmt.Sprintf("ORDER BY %s %s", orderBy, orderDir)
}

// 获取访问统计信息
func (m *MySQLClient) GetAccessStats(query *AccessRecordQuery) (*AccessStats, error) {
	whereClause, args := m.buildWhereClause(query)
	
	stats := &AccessStats{
		ActionStats: make(map[string]int64),
	}

	// 基础统计
	basicSQL := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_requests,
			COUNT(DISTINCT fingerprint) as unique_users,
			COUNT(DISTINCT ip) as unique_ips
		FROM access_logs 
		WHERE %s`, whereClause)
	
	err := m.db.QueryRow(basicSQL, args...).Scan(
		&stats.TotalRequests, &stats.UniqueUsers, &stats.UniqueIPs)
	if err != nil {
		return nil, fmt.Errorf("查询基础统计失败: %v", err)
	}

	// 动作统计
	actionSQL := fmt.Sprintf(`
		SELECT action, COUNT(*) as count 
		FROM access_logs 
		WHERE %s 
		GROUP BY action`, whereClause)
	
	rows, err := m.db.Query(actionSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("查询动作统计失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var action string
		var count int64
		if err := rows.Scan(&action, &count); err != nil {
			continue
		}
		stats.ActionStats[action] = count
	}

	// 小时统计
	hourlyStats, err := m.getHourlyStats(whereClause, args)
	if err == nil {
		stats.HourlyStats = hourlyStats
	}

	// 路径统计
	pathStats, err := m.getPathStats(whereClause, args)
	if err == nil {
		stats.PathStats = pathStats
	}

	// User-Agent统计
	userAgentStats, err := m.getUserAgentStats(whereClause, args)
	if err == nil {
		stats.UserAgentStats = userAgentStats
	}

	// 分数分布统计
	scoreDistribution, err := m.getScoreDistribution(whereClause, args)
	if err == nil {
		stats.ScoreDistribution = scoreDistribution
	}

	return stats, nil
}

// 获取小时统计
func (m *MySQLClient) getHourlyStats(whereClause string, args []interface{}) ([]HourlyAccessStat, error) {
	sql := fmt.Sprintf(`
		SELECT HOUR(timestamp) as hour, COUNT(*) as count 
		FROM access_logs 
		WHERE %s 
		GROUP BY HOUR(timestamp) 
		ORDER BY hour`, whereClause)

	rows, err := m.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []HourlyAccessStat
	for rows.Next() {
		var stat HourlyAccessStat
		if err := rows.Scan(&stat.Hour, &stat.Count); err != nil {
			continue
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

// 获取路径统计
func (m *MySQLClient) getPathStats(whereClause string, args []interface{}) ([]PathAccessStat, error) {
	sql := fmt.Sprintf(`
		SELECT path, COUNT(*) as count 
		FROM access_logs 
		WHERE %s 
		GROUP BY path 
		ORDER BY count DESC 
		LIMIT 20`, whereClause)

	rows, err := m.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []PathAccessStat
	for rows.Next() {
		var stat PathAccessStat
		if err := rows.Scan(&stat.Path, &stat.Count); err != nil {
			continue
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

// 获取User-Agent统计
func (m *MySQLClient) getUserAgentStats(whereClause string, args []interface{}) ([]UserAgentStat, error) {
	sql := fmt.Sprintf(`
		SELECT user_agent, COUNT(*) as count 
		FROM access_logs 
		WHERE %s AND user_agent IS NOT NULL AND user_agent != ''
		GROUP BY user_agent 
		ORDER BY count DESC 
		LIMIT 10`, whereClause)

	rows, err := m.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []UserAgentStat
	for rows.Next() {
		var stat UserAgentStat
		if err := rows.Scan(&stat.UserAgent, &stat.Count); err != nil {
			continue
		}
		// 截断过长的User-Agent
		if len(stat.UserAgent) > 100 {
			stat.UserAgent = stat.UserAgent[:100] + "..."
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

// 获取分数分布统计
func (m *MySQLClient) getScoreDistribution(whereClause string, args []interface{}) ([]ScoreDistributionStat, error) {
	sql := fmt.Sprintf(`
		SELECT 
			CASE 
				WHEN score >= 90 THEN '90-100'
				WHEN score >= 80 THEN '80-89'
				WHEN score >= 70 THEN '70-79'
				WHEN score >= 60 THEN '60-69'
				WHEN score >= 50 THEN '50-59'
				WHEN score >= 40 THEN '40-49'
				WHEN score >= 30 THEN '30-39'
				WHEN score >= 20 THEN '20-29'
				WHEN score >= 10 THEN '10-19'
				ELSE '0-9'
			END as score_range,
			COUNT(*) as count
		FROM access_logs 
		WHERE %s
		GROUP BY score_range 
		ORDER BY score_range DESC`, whereClause)

	rows, err := m.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []ScoreDistributionStat
	for rows.Next() {
		var stat ScoreDistributionStat
		if err := rows.Scan(&stat.ScoreRange, &stat.Count); err != nil {
			continue
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

// 根据用户指纹获取访问记录
func (m *MySQLClient) GetUserAccessRecords(fingerprint string, limit int, offset int) ([]AccessRecord, error) {
	query := &AccessRecordQuery{
		Fingerprint: fingerprint,
		Limit:       limit,
		Offset:      offset,
		OrderBy:     "timestamp",
		OrderDir:    "DESC",
	}
	
	result, err := m.QueryAccessRecords(query)
	if err != nil {
		return nil, err
	}
	
	return result.Records, nil
}

// 获取最近的访问记录
func (m *MySQLClient) GetRecentAccessRecords(minutes int, limit int) ([]AccessRecord, error) {
	startTime := time.Now().Add(-time.Duration(minutes) * time.Minute)
	
	query := &AccessRecordQuery{
		StartTime: startTime,
		Limit:     limit,
		Offset:    0,
		OrderBy:   "timestamp",
		OrderDir:  "DESC",
	}
	
	result, err := m.QueryAccessRecords(query)
	if err != nil {
		return nil, err
	}
	
	return result.Records, nil
}

// 删除过期的访问记录
func (m *MySQLClient) CleanupOldAccessRecords(days int) (int64, error) {
	cutoffTime := time.Now().AddDate(0, 0, -days)
	
	result, err := m.db.Exec("DELETE FROM access_logs WHERE timestamp < ?", cutoffTime)
	if err != nil {
		return 0, fmt.Errorf("删除过期记录失败: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取影响行数失败: %v", err)
	}
	
	return rowsAffected, nil
}
