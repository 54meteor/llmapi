package model

import (
	"one-api/common"
	"one-api/common/logger"
	"time"
)

type Group struct {
	Id        int64   `json:"id" gorm:"primaryKey"`
	Name      string  `json:"name" gorm:"uniqueIndex;size:32;not null"`
	Ratio     float64 `json:"ratio" gorm:"type:float;default:1.0"`
	CreatedAt int64   `json:"created_at"`
	UpdatedAt int64   `json:"updated_at"`
}

func GetAllGroups() ([]*Group, error) {
	var groups []*Group
	err := DB.Order("id").Find(&groups).Error
	return groups, err
}

func GetGroupById(id int64) (*Group, error) {
	group := Group{Id: id}
	err := DB.First(&group, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func GetGroupByName(name string) (*Group, error) {
	group := Group{Name: name}
	err := DB.First(&group, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (group *Group) Create() error {
	group.CreatedAt = time.Now().UnixMilli()
	group.UpdatedAt = time.Now().UnixMilli()
	return DB.Create(group).Error
}

func (group *Group) Update() error {
	group.UpdatedAt = time.Now().UnixMilli()
	return DB.Save(group).Error
}

func (group *Group) Delete() error {
	return DB.Delete(group).Error
}

func IsGroupUsedByUsers(groupName string) bool {
	var count int64
	DB.Model(&User{}).Where("`group` = ?", groupName).Count(&count)
	return count > 0
}

func EnsureDefaultGroup() {
	group, err := GetGroupByName("default")
	if err == nil && group != nil {
		return
	}
	defaultGroup := Group{
		Name:  "default",
		Ratio: 1.0,
	}
	if err := defaultGroup.Create(); err != nil {
		logger.SysError("failed to create default group: " + err.Error())
	}
}

func ReloadGroupRatioFromDB() {
	groups, err := GetAllGroups()
	if err != nil {
		logger.SysError("failed to load groups from database: " + err.Error())
		return
	}
	ratioMap := make(map[string]float64)
	for _, group := range groups {
		ratioMap[group.Name] = group.Ratio
	}
	common.SetGroupRatio(ratioMap)
}
