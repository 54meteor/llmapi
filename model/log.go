package model

import (
	"context"
	"fmt"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/helper"
	"one-api/common/logger"
	"time"

	"gorm.io/gorm"
)

type Log struct {
	Id               int    `json:"id"`
	UserId           int    `json:"user_id" gorm:"index"`
	CreatedAt        int64  `json:"created_at" gorm:"bigint;index:idx_created_at_type"`
	Type             int    `json:"type" gorm:"index:idx_created_at_type"`
	Content          string `json:"content"`
	Username         string `json:"username" gorm:"index:index_username_model_name,priority:2;default:''"`
	TokenName        string `json:"token_name" gorm:"index;default:''"`
	ModelName        string `json:"model_name" gorm:"index;index:index_username_model_name,priority:1;default:''"`
	Quota            int    `json:"quota" gorm:"default:0"`
	PromptTokens     int    `json:"prompt_tokens" gorm:"default:0"`
	CompletionTokens int    `json:"completion_tokens" gorm:"default:0"`
	ChannelId        int    `json:"channel" gorm:"index"`
}

const (
	LogTypeUnknown = iota
	LogTypeTopup
	LogTypeConsume
	LogTypeManage
	LogTypeSystem
)

func RecordLog(userId int, logType int, content string) {
	if logType == LogTypeConsume && !config.LogConsumeEnabled {
		return
	}
	log := &Log{
		UserId:    userId,
		Username:  GetUsernameById(userId),
		CreatedAt: helper.GetTimestamp(),
		Type:      logType,
		Content:   content,
	}
	err := DB.Create(log).Error
	if err != nil {
		logger.SysError("failed to record log: " + err.Error())
	}
}

func RecordConsumeLog(ctx context.Context, userId int, channelId int, promptTokens int, completionTokens int, modelName string, tokenName string, quota int, content string) {
	logger.Info(ctx, fmt.Sprintf("record consume log: userId=%d, channelId=%d, promptTokens=%d, completionTokens=%d, modelName=%s, tokenName=%s, quota=%d, content=%s", userId, channelId, promptTokens, completionTokens, modelName, tokenName, quota, content))
	if !config.LogConsumeEnabled {
		return
	}
	log := &Log{
		UserId:           userId,
		Username:         GetUsernameById(userId),
		CreatedAt:        helper.GetTimestamp(),
		Type:             LogTypeConsume,
		Content:          content,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TokenName:        tokenName,
		ModelName:        modelName,
		Quota:            quota,
		ChannelId:        channelId,
	}
	err := DB.Create(log).Error
	if err != nil {
		logger.Error(ctx, "failed to record log: "+err.Error())
	}
}

func GetAllLogs(logType int, startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string, startIdx int, num int, channel int) (logs []*Log, err error) {
	var tx *gorm.DB
	if logType == LogTypeUnknown {
		tx = DB
	} else {
		tx = DB.Where("type = ?", logType)
	}
	if modelName != "" {
		tx = tx.Where("model_name = ?", modelName)
	}
	if username != "" {
		tx = tx.Where("username = ?", username)
	}
	if tokenName != "" {
		tx = tx.Where("token_name = ?", tokenName)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}
	if channel != 0 {
		tx = tx.Where("channel_id = ?", channel)
	}
	err = tx.Order("id desc").Limit(num).Offset(startIdx).Find(&logs).Error
	return logs, err
}

func GetUserLogs(userId int, logType int, startTimestamp int64, endTimestamp int64, modelName string, tokenName string, startIdx int, num int) (logs []*Log, err error) {
	var tx *gorm.DB
	if logType == LogTypeUnknown {
		tx = DB.Where("user_id = ?", userId)
	} else {
		tx = DB.Where("user_id = ? and type = ?", userId, logType)
	}
	if modelName != "" {
		tx = tx.Where("model_name = ?", modelName)
	}
	if tokenName != "" {
		tx = tx.Where("token_name = ?", tokenName)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}
	err = tx.Order("id desc").Limit(num).Offset(startIdx).Omit("id").Find(&logs).Error
	return logs, err
}

func SearchAllLogs(keyword string) (logs []*Log, err error) {
	err = DB.Where("type = ? or content LIKE ?", keyword, keyword+"%").Order("id desc").Limit(config.MaxRecentItems).Find(&logs).Error
	return logs, err
}

func SearchUserLogs(userId int, keyword string) (logs []*Log, err error) {
	err = DB.Where("user_id = ? and type = ?", userId, keyword).Order("id desc").Limit(config.MaxRecentItems).Omit("id").Find(&logs).Error
	return logs, err
}

func SumUsedQuota(logType int, startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string, channel int) (quota int) {
	tx := DB.Table("logs").Select("ifnull(sum(quota),0)")
	if username != "" {
		tx = tx.Where("username = ?", username)
	}
	if tokenName != "" {
		tx = tx.Where("token_name = ?", tokenName)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}
	if modelName != "" {
		tx = tx.Where("model_name = ?", modelName)
	}
	if channel != 0 {
		tx = tx.Where("channel_id = ?", channel)
	}
	tx.Where("type = ?", LogTypeConsume).Scan(&quota)
	return quota
}

func SumUsedToken(logType int, startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string) (token int) {
	tx := DB.Table("logs").Select("ifnull(sum(prompt_tokens),0) + ifnull(sum(completion_tokens),0)")
	if username != "" {
		tx = tx.Where("username = ?", username)
	}
	if tokenName != "" {
		tx = tx.Where("token_name = ?", tokenName)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}
	if modelName != "" {
		tx = tx.Where("model_name = ?", modelName)
	}
	tx.Where("type = ?", LogTypeConsume).Scan(&token)
	return token
}

func DeleteOldLog(targetTimestamp int64) (int64, error) {
	result := DB.Where("created_at < ?", targetTimestamp).Delete(&Log{})
	return result.RowsAffected, result.Error
}

type LogStatistic struct {
	Day              string `gorm:"column:day"`
	ModelName        string `gorm:"column:model_name"`
	RequestCount     int    `gorm:"column:request_count"`
	Quota            int    `gorm:"column:quota"`
	PromptTokens     int    `gorm:"column:prompt_tokens"`
	CompletionTokens int    `gorm:"column:completion_tokens"`
}

func SearchLogsByDayAndModel(userId, start, end int) (LogStatistics []*LogStatistic, err error) {
	groupSelect := "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d') as day"

	if common.UsingPostgreSQL {
		groupSelect = "TO_CHAR(date_trunc('day', to_timestamp(created_at)), 'YYYY-MM-DD') as day"
	}

	if common.UsingSQLite {
		groupSelect = "strftime('%Y-%m-%d', datetime(created_at, 'unixepoch')) as day"
	}

	err = DB.Raw(`
		SELECT `+groupSelect+`,
		model_name, count(1) as request_count,
		sum(quota) as quota,
		sum(prompt_tokens) as prompt_tokens,
		sum(completion_tokens) as completion_tokens
		FROM logs
		WHERE type=2
		AND user_id= ?
		AND created_at BETWEEN ? AND ?
		GROUP BY day, model_name
		ORDER BY day, model_name
	`, userId, start, end).Scan(&LogStatistics).Error

	return LogStatistics, err
}

type DashboardStat struct {
	RequestCount   int64 `json:"request_count"`
	QuotaUsed      int64 `json:"quota_used"`
	ActiveUsers    int64 `json:"active_users"`
	ActiveChannels int64 `json:"active_channels"`
}

type TrendStat struct {
	Day          string `json:"day"`
	RequestCount int64  `json:"request_count"`
	Quota        int64  `json:"quota"`
}

type ModelStat struct {
	Model string  `json:"model"`
	Count int64   `json:"count"`
	Ratio float64 `json:"ratio"`
}

type ChannelStat struct {
	ChannelId       int     `json:"channel_id"`
	Name            string  `json:"name"`
	SuccessRate     float64 `json:"success_rate"`
	AvgResponseTime int     `json:"avg_response_time"`
	Balance         float64 `json:"balance"`
}

type UserStat struct {
	UserId       int    `json:"user_id"`
	Username     string `json:"username"`
	QuotaUsed    int64  `json:"quota_used"`
	RequestCount int64  `json:"request_count"`
}

func GetDashboardToday() (*DashboardStat, error) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).Unix()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, loc).Unix()

	var stat DashboardStat
	var err error

	err = DB.Raw(`
		SELECT
			COUNT(1) as request_count,
			COALESCE(SUM(quota), 0) as quota_used,
			COUNT(DISTINCT user_id) as active_users,
			COUNT(DISTINCT channel_id) as active_channels
		FROM logs
		WHERE type = ? AND created_at >= ? AND created_at <= ?
	`, LogTypeConsume, startOfDay, endOfDay).Scan(&stat).Error

	return &stat, err
}

func GetDashboardTrend7Days() ([]*TrendStat, error) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -6).Unix()

	groupSelect := "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d') as day"
	if common.UsingPostgreSQL {
		groupSelect = "TO_CHAR(date_trunc('day', to_timestamp(created_at)), 'YYYY-MM-DD') as day"
	}
	if common.UsingSQLite {
		groupSelect = "strftime('%Y-%m-%d', datetime(created_at, 'unixepoch')) as day"
	}

	var trends []*TrendStat
	err := DB.Raw(`
		SELECT `+groupSelect+`,
			COUNT(1) as request_count,
			COALESCE(SUM(quota), 0) as quota
		FROM logs
		WHERE type = ? AND created_at >= ?
		GROUP BY day
		ORDER BY day ASC
	`, LogTypeConsume, startOfDay).Scan(&trends).Error

	return trends, err
}

func GetDashboardModelDistribution() ([]*ModelStat, error) {
	var total int64
	err := DB.Model(&Log{}).Where("type = ?", LogTypeConsume).Count(&total).Error
	if err != nil || total == 0 {
		return nil, err
	}

	var stats []*ModelStat
	err = DB.Raw(`
		SELECT model_name as model, COUNT(1) as count
		FROM logs
		WHERE type = ?
		GROUP BY model_name
		ORDER BY count DESC
		LIMIT 10
	`, LogTypeConsume).Scan(&stats).Error
	if err != nil {
		return nil, err
	}

	for _, stat := range stats {
		stat.Ratio = float64(stat.Count) / float64(total)
	}

	return stats, err
}

func GetDashboardChannelHealth() ([]*ChannelStat, error) {
	var channels []*Channel
	err := DB.Select("id", "name", "response_time", "balance").Find(&channels).Error
	if err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).Unix()

	type channelLogStat struct {
		ChannelId  int `json:"channel_id"`
		RequestCnt int `json:"request_cnt"`
	}
	var logStats []channelLogStat
	err = DB.Raw(`
		SELECT channel_id as channel_id, COUNT(1) as request_cnt
		FROM logs
		WHERE type = ? AND created_at >= ?
		GROUP BY channel_id
	`, LogTypeConsume, startOfDay).Scan(&logStats).Error
	if err != nil {
		return nil, err
	}

	totalByChannel := make(map[int]int)
	for _, ls := range logStats {
		totalByChannel[ls.ChannelId] = ls.RequestCnt
	}

	var stats []*ChannelStat
	for _, ch := range channels {
		stat := &ChannelStat{
			ChannelId:       ch.Id,
			Name:            ch.Name,
			AvgResponseTime: ch.ResponseTime,
			Balance:         ch.Balance,
			SuccessRate:     1.0,
		}
		if cnt, ok := totalByChannel[ch.Id]; ok && cnt > 0 {
			stat.SuccessRate = 1.0
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func GetDashboardTopUsers(limit int) ([]*UserStat, error) {
	var stats []*UserStat
	err := DB.Raw(`
		SELECT user_id, username,
			COALESCE(SUM(quota), 0) as quota_used,
			COUNT(1) as request_count
		FROM logs
		WHERE type = ?
		GROUP BY user_id, username
		ORDER BY quota_used DESC
		LIMIT ?
	`, LogTypeConsume, limit).Scan(&stats).Error

	return stats, err
}

type UserUsageStat struct {
	TotalUsed      int64        `json:"total_used"`
	TotalRequests  int64        `json:"total_requests"`
	QuotaRemaining int          `json:"quota_remaining"`
	QuotaPercent   float64      `json:"quota_percent"`
	Trend7Days     []*TrendStat `json:"trend_7days"`
}

func GetUserUsageStat(userId int) (*UserUsageStat, error) {
	user, err := GetUserById(userId, false)
	if err != nil {
		return nil, err
	}

	var totalUsed int64
	err = DB.Model(&Log{}).Where("user_id = ? AND type = ?", userId, LogTypeConsume).Select("COALESCE(SUM(quota), 0)").Scan(&totalUsed).Error
	if err != nil {
		return nil, err
	}

	var totalRequests int64
	err = DB.Model(&Log{}).Where("user_id = ? AND type = ?", userId, LogTypeConsume).Count(&totalRequests).Error
	if err != nil {
		return nil, err
	}

	quotaRemaining := user.Quota
	quotaPercent := 0.0
	if user.Quota > 0 {
		quotaPercent = float64(totalUsed) / float64(user.Quota)
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -6).Unix()

	groupSelect := "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d') as day"
	if common.UsingPostgreSQL {
		groupSelect = "TO_CHAR(date_trunc('day', to_timestamp(created_at)), 'YYYY-MM-DD') as day"
	}
	if common.UsingSQLite {
		groupSelect = "strftime('%Y-%m-%d', datetime(created_at, 'unixepoch')) as day"
	}

	var trends []*TrendStat
	err = DB.Raw(`
		SELECT `+groupSelect+`,
			COUNT(1) as request_count,
			COALESCE(SUM(quota), 0) as quota
		FROM logs
		WHERE type = ? AND user_id = ? AND created_at >= ?
		GROUP BY day
		ORDER BY day ASC
	`, LogTypeConsume, userId, startOfDay).Scan(&trends).Error
	if err != nil {
		return nil, err
	}

	stat := &UserUsageStat{
		TotalUsed:      totalUsed,
		TotalRequests:  totalRequests,
		QuotaRemaining: quotaRemaining,
		QuotaPercent:   quotaPercent,
		Trend7Days:     trends,
	}

	return stat, nil
}
