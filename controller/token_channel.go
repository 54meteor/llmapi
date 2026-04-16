package controller

import (
	"net/http"
	"one-api/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetTokenChannels(c *gin.Context) {
	tokenId, err := strconv.Atoi(c.Param("token_id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "invalid token_id",
		})
		return
	}

	channels, err := model.GetTokenChannelWithDetails(tokenId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	for i := range channels {
		channel := channels[i]
		quotaLimit := channel["quota_limit"].(int64)
		usedQuota := channel["used_quota"].(int64)
		remainQuota := quotaLimit - usedQuota
		if quotaLimit == 0 {
			channel["remain_quota"] = "不限"
			channel["remain_percent"] = 100
		} else {
			channel["remain_quota"] = remainQuota
			channel["remain_percent"] = int(float64(remainQuota) / float64(quotaLimit) * 100)
		}
		channels[i] = channel
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    channels,
	})
}

func CreateTokenChannel(c *gin.Context) {
	tokenChannel := model.TokenChannel{}
	err := c.ShouldBindJSON(&tokenChannel)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if tokenChannel.TokenId == 0 || tokenChannel.ChannelId == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "token_id and channel_id are required",
		})
		return
	}

	existing, err := model.GetTokenChannel(tokenChannel.TokenId, tokenChannel.ChannelId)
	if err == nil && existing != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该渠道已绑定到此令牌",
		})
		return
	}

	err = tokenChannel.Create()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    tokenChannel,
	})
}

func UpdateTokenChannel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "invalid id",
		})
		return
	}

	tokenChannel, err := model.GetTokenChannelById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	updateData := model.TokenChannel{}
	err = c.ShouldBindJSON(&updateData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if updateData.Priority > 0 {
		tokenChannel.Priority = updateData.Priority
	}
	if updateData.QuotaLimit > 0 || updateData.QuotaLimit == 0 {
		tokenChannel.QuotaLimit = updateData.QuotaLimit
	}

	err = tokenChannel.Update()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    tokenChannel,
	})
}

func DeleteTokenChannel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "invalid id",
		})
		return
	}

	err = model.DeleteTokenChannelById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}
