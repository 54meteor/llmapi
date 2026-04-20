package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/helper"
	"one-api/model"
	"strconv"
)

func GetAllTokens(c *gin.Context) {
	userId := c.GetInt("id")
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	tokens, err := model.GetAllUserTokens(userId, p*config.ItemsPerPage, config.ItemsPerPage)
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
		"data":    tokens,
	})
	return
}

func SearchTokens(c *gin.Context) {
	keyword := c.Query("keyword")
	queryUserId := c.Query("user_id")

	var userId int
	if c.GetInt("role") >= common.RoleAdminUser && queryUserId != "" {
		if uid, err := strconv.Atoi(queryUserId); err == nil {
			userId = uid
		} else {
			userId = c.GetInt("id")
		}
	} else {
		userId = c.GetInt("id")
	}
	tokens, err := model.SearchUserTokens(userId, keyword)
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
		"data":    tokens,
	})
	return
}

func GetToken(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	userId := c.GetInt("id")
	isAdmin := c.GetInt("role") >= common.RoleAdminUser
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	var token *model.Token
	if isAdmin {
		token, err = model.GetTokenById(id)
	} else {
		token, err = model.GetTokenByIds(id, userId)
	}
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
		"data":    token,
	})
	return
}

func GetTokenStatus(c *gin.Context) {
	tokenId := c.GetInt("token_id")
	userId := c.GetInt("id")
	token, err := model.GetTokenByIds(tokenId, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	expiredAt := token.ExpiredTime
	if expiredAt == -1 {
		expiredAt = 0
	}
	c.JSON(http.StatusOK, gin.H{
		"object":          "credit_summary",
		"total_granted":   token.RemainQuota,
		"total_used":      0, // not supported currently
		"total_available": token.RemainQuota,
		"expires_at":      expiredAt * 1000,
	})
}

func AddToken(c *gin.Context) {
	token := model.Token{}
	err := c.ShouldBindJSON(&token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if token.Name == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "令牌名称不能为空",
		})
		return
	}
	if len(token.Name) > 30 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "令牌名称过长",
		})
		return
	}
	userId := c.GetInt("id")
	if c.GetInt("role") >= common.RoleAdminUser {
		if uidStr := c.Query("user_id"); uidStr != "" {
			if uid, err := strconv.Atoi(uidStr); err == nil {
				userId = uid
			}
		}
	}
	cleanToken := model.Token{
		UserId:              userId,
		Name:                token.Name,
		Key:                 helper.GenerateKey(),
		CreatedTime:         helper.GetTimestamp(),
		AccessedTime:        helper.GetTimestamp(),
		ExpiredTime:         token.ExpiredTime,
		RemainQuota:         token.RemainQuota,
		UnlimitedQuota:      token.UnlimitedQuota,
		SwitchThreshold:     token.SwitchThreshold,
		SwitchThresholdType: token.SwitchThresholdType,
		AlertThreshold:      token.AlertThreshold,
		AlertThresholdType:  token.AlertThresholdType,
		SmartChannelEnabled: token.SmartChannelEnabled,
	}
	err = cleanToken.Insert()
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
	return
}

func DeleteToken(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userId := c.GetInt("id")
	err := model.DeleteTokenById(id, userId)
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
	return
}

func UpdateToken(c *gin.Context) {
	userId := c.GetInt("id")
	isAdmin := c.GetInt("role") >= common.RoleAdminUser
	statusOnly := c.Query("status_only")
	token := model.Token{}
	err := c.ShouldBindJSON(&token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if len(token.Name) > 30 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "令牌名称过长",
		})
		return
	}

	var cleanToken *model.Token
	if isAdmin && token.UserId > 0 {
		cleanToken, err = model.GetTokenById(token.Id)
	} else {
		cleanToken, err = model.GetTokenByIds(token.Id, userId)
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if token.Status == common.TokenStatusEnabled {
		if cleanToken.Status == common.TokenStatusExpired && cleanToken.ExpiredTime <= helper.GetTimestamp() && cleanToken.ExpiredTime != -1 {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "令牌已过期，无法启用，请先修改令牌过期时间，或者设置为永不过期",
			})
			return
		}
		if cleanToken.Status == common.TokenStatusExhausted && cleanToken.RemainQuota <= 0 && !cleanToken.UnlimitedQuota {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "令牌可用额度已用尽，无法启用，请先修改令牌剩余额度，或者设置为无限额度",
			})
			return
		}
	}
	if statusOnly != "" {
		cleanToken.Status = token.Status
	} else {
		cleanToken.Name = token.Name
		cleanToken.ExpiredTime = token.ExpiredTime
		cleanToken.RemainQuota = token.RemainQuota
		cleanToken.UnlimitedQuota = token.UnlimitedQuota
		cleanToken.SwitchThreshold = token.SwitchThreshold
		cleanToken.SwitchThresholdType = token.SwitchThresholdType
		cleanToken.AlertThreshold = token.AlertThreshold
		cleanToken.AlertThresholdType = token.AlertThresholdType
		cleanToken.SmartChannelEnabled = token.SmartChannelEnabled
		if isAdmin && token.UserId > 0 {
			cleanToken.UserId = token.UserId
		}
	}
	err = cleanToken.Update()
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
		"data":    cleanToken,
	})
	return
}

type TokenWithUsername struct {
	model.Token
	Username string `json:"username"`
}

func GetAllTokensAdmin(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	filterUserId := 0
	if uidStr := c.Query("user_id"); uidStr != "" {
		if uid, err := strconv.Atoi(uidStr); err == nil {
			filterUserId = uid
		}
	}
	tokens, err := model.GetAllTokensAdmin(filterUserId, p*config.ItemsPerPage, config.ItemsPerPage)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	var result []TokenWithUsername
	for _, token := range tokens {
		username := model.GetUsernameById(token.UserId)
		result = append(result, TokenWithUsername{
			Token:    *token,
			Username: username,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    result,
	})
	return
}
