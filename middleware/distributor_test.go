package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"one-api/common"
	"one-api/model"
)

func TestMain(m *testing.M) {
	// 初始化数据库连接
	if err := model.InitDB(); err != nil {
		panic("failed to init DB: " + err.Error())
	}
	// 初始化 Redis 连接
	if err := common.InitRedisClient(); err != nil {
		panic("failed to init Redis: " + err.Error())
	}
	os.Exit(m.Run())
}

func TestDistributeWithoutChannelId(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/chat/completions", nil)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("id", 1)

	Distribute()(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "该令牌未绑定任何渠道")
}

func TestDistributeWithChannelIdInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/chat/completions", nil)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("id", 1)
	c.Set("channelId", "invalid")

	Distribute()(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "无效的渠道 Id")
}
