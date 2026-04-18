package util

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/logger"
	"one-api/model"
	"one-api/relay/channel/openai"
	"strings"
)

type RetryOption struct {
	TokenId   int
	ModelName string
	Ctx       context.Context
}

type RetryResponse struct {
	Response  *http.Response
	ChannelId int
	Error     *openai.ErrorWithStatusCode
}

func DoRequestWithRetry(meta *RelayMeta, requestBody io.Reader, apiType int, opt *RetryOption) *RetryResponse {
	retryTimes := config.RetryTimes
	if retryTimes <= 0 {
		return doRequestOnce(meta, requestBody, apiType)
	}

	var lastResp *http.Response
	var lastErr *openai.ErrorWithStatusCode
	usedChannelIds := make(map[int]bool)
	usedChannelIds[meta.ChannelId] = true

	for attempt := 0; attempt <= retryTimes; attempt++ {
		if attempt > 0 {
			logger.Info(opt.Ctx, fmt.Sprintf("重试第 %d/%d 次，尝试切换渠道", attempt, retryTimes))
		}

		resp := doRequest(meta, requestBody, apiType)
		if resp.Error == nil && resp.Response != nil && resp.Response.StatusCode == http.StatusOK {
			return resp
		}

		if resp.Response != nil {
			resp.Response.Body.Close()
		}

		lastResp = nil
		lastErr = resp.Error

		if opt != nil && opt.TokenId > 0 && attempt < retryTimes {
			nextChannel, err := getNextChannel(opt.TokenId, opt.ModelName, usedChannelIds)
			if err != nil {
				logger.Info(opt.Ctx, fmt.Sprintf("无法获取下一个可用渠道: %s", err.Error()))
				break
			}

			if nextChannel == nil {
				logger.Info(opt.Ctx, "没有更多可用渠道")
				break
			}

			usedChannelIds[nextChannel.Id] = true
			meta.ChannelId = nextChannel.Id
			meta.ChannelType = nextChannel.Type
			meta.BaseURL = nextChannel.GetBaseURL()
			meta.APIVersion = nextChannel.Other
			meta.BillingMode = nextChannel.BillingMode
			meta.CountRatio = nextChannel.CountRatio

			apiKey := nextChannel.Key
			if nextChannel.Type == common.ChannelTypeOpenAIOAuth {
				apiKey = getOpenAIAccessToken(apiKey)
			}
			meta.APIKey = apiKey

			continue
		}

		break
	}

	return &RetryResponse{
		Response:  lastResp,
		ChannelId: meta.ChannelId,
		Error:     lastErr,
	}
}

func doRequestOnce(meta *RelayMeta, requestBody io.Reader, apiType int) *RetryResponse {
	return doRequest(meta, requestBody, apiType)
}

func doRequest(meta *RelayMeta, requestBody io.Reader, apiType int) *RetryResponse {
	fullRequestURL, err := GetRequestURLForRetry(meta, apiType)
	if err != nil {
		return &RetryResponse{Error: openai.ErrorWrapper(err, "get_request_url_failed", http.StatusInternalServerError)}
	}

	req, err := http.NewRequest("POST", fullRequestURL, requestBody)
	if err != nil {
		return &RetryResponse{Error: openai.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)}
	}

	setupRetryRequestHeaders(req, meta)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return &RetryResponse{Error: openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)}
	}

	if resp.StatusCode != http.StatusOK {
		return &RetryResponse{Response: resp, Error: openai.ErrorWrapper(fmt.Errorf("status code: %d", resp.StatusCode), "request_failed", resp.StatusCode)}
	}

	return &RetryResponse{Response: resp, ChannelId: meta.ChannelId}
}

func GetRequestURLForRetry(meta *RelayMeta, apiType int) (string, error) {
	baseURL := meta.BaseURL
	if baseURL == "" {
		baseURL = common.ChannelBaseURLs[meta.ChannelType]
	}

	switch meta.ChannelType {
	case common.ChannelTypeOpenAI, common.ChannelTypeOpenAIOAuth:
		return baseURL + "/v1/chat/completions", nil
	case common.ChannelTypeAzure:
		apiVersion := meta.APIVersion
		deploymentName := meta.Config["deployment_name"]
		if deploymentName == "" {
			deploymentName = "gpt-35-turbo"
		}
		return fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s", baseURL, deploymentName, apiVersion), nil
	case common.ChannelTypeXunfei:
		return baseURL + "/v3.5/chat", nil
	default:
		return baseURL + "/v1/chat/completions", nil
	}
}

func setupRetryRequestHeaders(req *http.Request, meta *RelayMeta) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", meta.APIKey))
	req.Header.Set("Content-Type", "application/json")
	if meta.ChannelType == common.ChannelTypeAzure {
		req.Header.Set("api-key", meta.APIKey)
	}
}

func getNextChannel(tokenId int, modelName string, usedIds map[int]bool) (*model.Channel, error) {
	if tokenId <= 0 {
		return nil, fmt.Errorf("token id is 0")
	}

	channels, err := model.GetTokenChannels(tokenId)
	if err != nil {
		return nil, err
	}

	if len(channels) == 0 {
		return nil, fmt.Errorf("token has no bound channels")
	}

	for _, tc := range channels {
		if usedIds[tc.ChannelId] {
			continue
		}

		remaining := tc.GetRemainingPercent()
		if remaining <= 0 {
			continue
		}

		channel, err := model.GetChannelById(tc.ChannelId, true)
		if err != nil {
			continue
		}

		if channel.Status != common.ChannelStatusEnabled {
			continue
		}

		return channel, nil
	}

	return nil, fmt.Errorf("no available channel")
}

type OpenAIOAuthToken struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
}

func getOpenAIAccessToken(key string) string {
	parts := strings.Split(key, ".")
	if len(parts) != 3 {
		return key
	}
	padding := 4 - len(parts[1])%4
	if padding != 4 {
		parts[1] += strings.Repeat("=", padding)
	}
	payload, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return key
	}
	var token OpenAIOAuthToken
	if err := json.Unmarshal(payload, &token); err != nil {
		return key
	}
	return token.AccessToken
}
