package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/common/logger"
	"one-api/model"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type OpenAIOAuthToken struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
}

func getOpenAIAccessToken(key string) string {
	parts := strings.Split(key, ".")
	if len(parts) != 3 {
		logger.SysError("invalid OpenAI OAuth key format")
		return key
	}
	padding := 4 - len(parts[1])%4
	if padding != 4 {
		parts[1] += strings.Repeat("=", padding)
	}
	payload, err := jwt.DecodeSegment(parts[1])
	if err != nil {
		logger.SysError("failed to decode JWT payload: " + err.Error())
		return key
	}
	var token OpenAIOAuthToken
	if err := json.Unmarshal(payload, &token); err != nil {
		logger.SysError("failed to parse OAuth token: " + err.Error())
		return key
	}
	return token.AccessToken
}

type ModelRequest struct {
	Model string `json:"model"`
}

func Distribute() func(c *gin.Context) {
	return func(c *gin.Context) {
		userId := c.GetInt("id")
		userGroup, _ := model.CacheGetUserGroup(userId)
		c.Set("group", userGroup)
		var channel *model.Channel
		channelId, ok := c.Get("channelId")
		if ok {
			id, err := strconv.Atoi(channelId.(string))
			if err != nil {
				abortWithMessage(c, http.StatusBadRequest, "无效的渠道 Id")
				return
			}
			channel, err = model.GetChannelById(id, true)
			if err != nil {
				abortWithMessage(c, http.StatusBadRequest, "无效的渠道 Id")
				return
			}
			if channel.Status != common.ChannelStatusEnabled {
				abortWithMessage(c, http.StatusForbidden, "该渠道已被禁用")
				return
			}
		} else {
			tokenId := c.GetInt("token_id")
			if tokenId > 0 {
				var modelRequest ModelRequest
				err := common.UnmarshalBodyReusable(c, &modelRequest)
				if err != nil {
					abortWithMessage(c, http.StatusBadRequest, "无效的请求")
					return
				}
				modelName := modelRequest.Model
				if modelName == "" {
					modelName = c.Param("model")
				}
				channel, err = model.SelectChannelByToken(tokenId, modelName)
				if err != nil {
					abortWithMessage(c, http.StatusServiceUnavailable, err.Error())
					return
				}
			} else {
				abortWithMessage(c, http.StatusForbidden, "该令牌未绑定任何渠道，请先在令牌管理中绑定渠道路由")
				return
			}
		}
		c.Set("channel", channel.Type)
		c.Set("channel_id", channel.Id)
		c.Set("channel_name", channel.Name)
		c.Set("model_mapping", channel.GetModelMapping())
		apiKey := channel.Key
		if channel.Type == common.ChannelTypeOpenAIOAuth {
			apiKey = getOpenAIAccessToken(apiKey)
		}
		c.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
		c.Set("base_url", channel.GetBaseURL())
		switch channel.Type {
		case common.ChannelTypeAzure:
			c.Set("api_version", channel.Other)
		case common.ChannelTypeXunfei:
			c.Set("api_version", channel.Other)
		case common.ChannelTypeGemini:
			c.Set("api_version", channel.Other)
		case common.ChannelTypeAIProxyLibrary:
			c.Set("library_id", channel.Other)
		case common.ChannelTypeAli:
			c.Set("plugin", channel.Other)
		}
		c.Set("billing_mode", channel.BillingMode)
		c.Set("count_ratio", channel.CountRatio)
		c.Next()
	}
}
