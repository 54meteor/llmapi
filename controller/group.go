package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"one-api/model"
	"strconv"
)

func GetGroups(c *gin.Context) {
	groups, err := model.GetAllGroups()
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
		"data":    groups,
	})
}

func GetGroup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "invalid id",
		})
		return
	}
	group, err := model.GetGroupById(id)
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
		"data":    group,
	})
}

func CreateGroup(c *gin.Context) {
	group := model.Group{}
	err := c.ShouldBindJSON(&group)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if group.Name == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "分组名称不能为空",
		})
		return
	}
	existingGroup, _ := model.GetGroupByName(group.Name)
	if existingGroup != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "分组名称已存在",
		})
		return
	}
	if group.Ratio == 0 {
		group.Ratio = 1.0
	}
	err = group.Create()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	model.ReloadGroupRatioFromDB()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    group,
	})
}

func UpdateGroup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "invalid id",
		})
		return
	}
	group, err := model.GetGroupById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	updateData := model.Group{}
	err = c.ShouldBindJSON(&updateData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if updateData.Name != "" && updateData.Name != group.Name {
		existingGroup, _ := model.GetGroupByName(updateData.Name)
		if existingGroup != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "分组名称已存在",
			})
			return
		}
		group.Name = updateData.Name
	}
	if updateData.Ratio != 0 {
		group.Ratio = updateData.Ratio
	}
	err = group.Update()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	model.ReloadGroupRatioFromDB()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    group,
	})
}

func DeleteGroup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "invalid id",
		})
		return
	}
	group, err := model.GetGroupById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if group.Name == "default" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "默认分组不可删除",
		})
		return
	}
	if model.IsGroupUsedByUsers(group.Name) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该分组下仍有用户，无法删除",
		})
		return
	}
	err = group.Delete()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	model.ReloadGroupRatioFromDB()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}
