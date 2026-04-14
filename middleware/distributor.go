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
			// Select a channel for the user
			var modelRequest ModelRequest
			err := common.UnmarshalBodyReusable(c, &modelRequest)
			if err != nil {
				abortWithMessage(c, http.StatusBadRequest, "无效的请求")
				return
			}
			if strings.HasPrefix(c.Request.URL.Path, "/v1/moderations") {
				if modelRequest.Model == "" {
					modelRequest.Model = "text-moderation-stable"
				}
			}
			if strings.HasSuffix(c.Request.URL.Path, "embeddings") {
				if modelRequest.Model == "" {
					modelRequest.Model = c.Param("model")
				}
			}
			if strings.HasPrefix(c.Request.URL.Path, "/v1/images/generations") {
				if modelRequest.Model == "" {
					modelRequest.Model = "dall-e-2"
				}
			}
			if strings.HasPrefix(c.Request.URL.Path, "/v1/audio/transcriptions") || strings.HasPrefix(c.Request.URL.Path, "/v1/audio/translations") {
				if modelRequest.Model == "" {
					modelRequest.Model = "whisper-1"
				}
			}
			channel, err = model.CacheGetRandomSatisfiedChannel(userGroup, modelRequest.Model)
			if err != nil {
				message := fmt.Sprintf("当前分组 %s 下对于模型 %s 无可用渠道", userGroup, modelRequest.Model)
				if channel != nil {
					logger.SysError(fmt.Sprintf("渠道不存在：%d", channel.Id))
					message = "数据库一致性已被破坏，请联系管理员"
				}
				abortWithMessage(c, http.StatusServiceUnavailable, message)
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
		c.Next()
	}
}
