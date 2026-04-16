package model

import "one-api/common/helper"

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
