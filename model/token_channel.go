package model

import (
	"errors"
	"fmt"
	"one-api/common"
	"one-api/common/helper"
	"one-api/common/logger"

	"gorm.io/gorm"
)

type TokenChannel struct {
	Id          int   `json:"id" gorm:"primaryKey"`
	TokenId     int   `json:"token_id" gorm:"index;not null"`
	ChannelId   int   `json:"channel_id" gorm:"index;not null"`
	Priority    int   `json:"priority" gorm:"default:1"`
	QuotaLimit  int64 `json:"quota_limit" gorm:"default:0"`
	UsedQuota   int64 `json:"used_quota" gorm:"default:0"`
	CreatedTime int64 `json:"created_time" gorm:"bigint"`
	UpdatedTime int64 `json:"updated_time" gorm:"bigint"`
}

func GetTokenChannels(tokenId int) ([]*TokenChannel, error) {
	var tokenChannels []*TokenChannel
	err := DB.Where("token_id = ?", tokenId).Order("priority ASC").Find(&tokenChannels).Error
	return tokenChannels, err
}

func GetTokenChannelById(id int) (*TokenChannel, error) {
	tokenChannel := TokenChannel{Id: id}
	err := DB.First(&tokenChannel, "id = ?", id).Error
	return &tokenChannel, err
}

func GetTokenChannel(tokenId int, channelId int) (*TokenChannel, error) {
	tokenChannel := TokenChannel{}
	err := DB.Where("token_id = ? AND channel_id = ?", tokenId, channelId).First(&tokenChannel).Error
	return &tokenChannel, err
}

func (tc *TokenChannel) Create() error {
	tc.CreatedTime = helper.GetTimestamp()
	tc.UpdatedTime = helper.GetTimestamp()
	return DB.Create(tc).Error
}

func (tc *TokenChannel) Update() error {
	tc.UpdatedTime = helper.GetTimestamp()
	return DB.Save(tc).Error
}

func (tc *TokenChannel) Delete() error {
	return DB.Delete(tc).Error
}

func DeleteTokenChannelById(id int) error {
	return DB.Where("id = ?", id).Delete(&TokenChannel{}).Error
}

func DeleteTokenChannelsByTokenId(tokenId int) error {
	return DB.Where("token_id = ?", tokenId).Delete(&TokenChannel{}).Error
}

func GetTokenChannelWithDetails(tokenId int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := DB.Table("token_channels").
		Select("token_channels.id, token_channels.token_id, token_channels.channel_id, channels.name as channel_name, channels.type as channel_type, token_channels.priority, token_channels.quota_limit, token_channels.used_quota").
		Joins("JOIN channels ON channels.id = token_channels.channel_id").
		Where("token_channels.token_id = ?", tokenId).
		Order("token_channels.priority ASC").
		Scan(&results).Error
	return results, err
}

func (tc *TokenChannel) GetRemainingPercent() int {
	if tc.QuotaLimit <= 0 {
		return 100
	}
	remain := tc.QuotaLimit - tc.UsedQuota
	if remain < 0 {
		remain = 0
	}
	return int(float64(remain) / float64(tc.QuotaLimit) * 100)
}

func (tc *TokenChannel) GetRemainingQuota() int64 {
	if tc.QuotaLimit <= 0 {
		return -1
	}
	remain := tc.QuotaLimit - tc.UsedQuota
	if remain < 0 {
		remain = 0
	}
	return remain
}

func UpdateTokenChannelUsedQuota(tokenId int, channelId int, quota int) error {
	return DB.Model(&TokenChannel{}).
		Where("token_id = ? AND channel_id = ?", tokenId, channelId).
		Update("used_quota", gorm.Expr("used_quota + ?", quota)).Error
}

func SelectChannelByToken(tokenId int, modelName string) (*Channel, error) {
	tokenChannels, err := GetTokenChannels(tokenId)
	if err != nil {
		return nil, err
	}

	if len(tokenChannels) == 0 {
		return nil, errors.New("该令牌未绑定任何渠道")
	}

	token, err := GetTokenById(tokenId)
	if err != nil {
		return nil, err
	}

	if !token.SmartChannelEnabled {
		return GetChannelById(tokenChannels[0].ChannelId, true)
	}

	alertThreshold := token.AlertThreshold
	if token.AlertThresholdType == "percent" {
		allBelowAlert := true
		for _, tc := range tokenChannels {
			if tc.GetRemainingPercent() >= alertThreshold {
				allBelowAlert = false
				break
			}
		}
		if allBelowAlert {
			logger.SysLog(fmt.Sprintf("Token %d 所有渠道额度低于 %d%%", tokenId, alertThreshold))
		}
	}

	switchThreshold := token.SwitchThreshold

	for _, tc := range tokenChannels {
		remaining := tc.GetRemainingPercent()

		if remaining < switchThreshold {
			continue
		}

		channel, err := GetChannelById(tc.ChannelId, true)
		if err != nil {
			continue
		}

		if channel.Status != common.ChannelStatusEnabled {
			continue
		}

		return channel, nil
	}

	return GetChannelById(tokenChannels[0].ChannelId, true)
}
