package common

import (
	"encoding/json"
	"one-api/common/logger"
	"sync"
)

var (
	GroupRatio     = make(map[string]float64)
	GroupRatioLock sync.RWMutex
)

func GroupRatio2JSONString() string {
	GroupRatioLock.RLock()
	defer GroupRatioLock.RUnlock()
	jsonBytes, err := json.Marshal(GroupRatio)
	if err != nil {
		logger.SysError("error marshalling group ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateGroupRatioByJSONString(jsonStr string) error {
	GroupRatioLock.Lock()
	defer GroupRatioLock.Unlock()
	GroupRatio = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &GroupRatio)
}

func GetGroupRatio(name string) float64 {
	GroupRatioLock.RLock()
	defer GroupRatioLock.RUnlock()
	ratio, ok := GroupRatio[name]
	if !ok {
		logger.SysError("group ratio not found: " + name)
		return 1
	}
	return ratio
}

func SetGroupRatio(ratioMap map[string]float64) {
	GroupRatioLock.Lock()
	defer GroupRatioLock.Unlock()
	GroupRatio = ratioMap
}
