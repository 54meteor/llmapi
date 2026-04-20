package controller

import (
	"bytes"
	"encoding/json"
	"io"
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
	if err := model.InitDB(); err != nil {
		panic("failed to init DB: " + err.Error())
	}
	if err := common.InitRedisClient(); err != nil {
		panic("failed to init Redis: " + err.Error())
	}
	os.Exit(m.Run())
}

func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/token", nil)
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

func TestAddTokenAsAdminWithUserId(t *testing.T) {
	c, w := setupTestContext()
	c.Set("id", 1)
	c.Set("role", common.RoleAdminUser)
	c.Request.URL.RawQuery = "user_id=2"

	tokenBody := map[string]interface{}{
		"name":            "test-token-admin",
		"expired_time":    -1,
		"remain_quota":    1000,
		"unlimited_quota": false,
	}
	jsonBody, _ := json.Marshal(tokenBody)
	c.Request.Body = io.NopCloser(bytes.NewReader(jsonBody))

	AddToken(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["success"])
}

func TestAddTokenAsCommonUserForcesOwnUserId(t *testing.T) {
	c, w := setupTestContext()
	c.Set("id", 3)
	c.Set("role", common.RoleCommonUser)
	c.Request.URL.RawQuery = "user_id=2"

	tokenBody := map[string]interface{}{
		"name":            "test-token-user",
		"expired_time":    -1,
		"remain_quota":    1000,
		"unlimited_quota": false,
	}
	jsonBody, _ := json.Marshal(tokenBody)
	c.Request.Body = io.NopCloser(bytes.NewReader(jsonBody))

	AddToken(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["success"])
}

func TestSearchTokensAsAdminWithUserIdFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/token/search?keyword=test&user_id=2", nil)
	c.Set("id", 1)
	c.Set("role", common.RoleAdminUser)

	SearchTokens(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["success"])
}

func TestSearchTokensAsCommonUserCannotFilterByOtherUserId(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/token/search?keyword=test&user_id=2", nil)
	c.Set("id", 3)
	c.Set("role", common.RoleCommonUser)

	SearchTokens(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["success"])
}

func TestGetAllTokensAdminReturnsUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/token/list-all?p=0", nil)
	c.Set("id", 1)
	c.Set("role", common.RoleAdminUser)

	GetAllTokensAdmin(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["success"])
}

func TestGetAllTokensAdminWithUserIdFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/token/list-all?user_id=2&p=0", nil)
	c.Set("id", 1)
	c.Set("role", common.RoleAdminUser)

	GetAllTokensAdmin(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["success"])
}

func TestUpdateTokenAsAdminCanModifyUserId(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/token", nil)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("id", 1)
	c.Set("role", common.RoleAdminUser)

	tokenBody := map[string]interface{}{
		"id":              1,
		"name":            "updated-token",
		"expired_time":    -1,
		"remain_quota":    2000,
		"user_id":         5,
	}
	jsonBody, _ := json.Marshal(tokenBody)
	c.Request.Body = io.NopCloser(bytes.NewReader(jsonBody))

	UpdateToken(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["success"] == true {
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, float64(5), data["user_id"])
	}
}

func TestUpdateTokenAsCommonUserCannotModifyUserId(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/token", nil)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("id", 1)
	c.Set("role", common.RoleCommonUser)

	tokenBody := map[string]interface{}{
		"id":              1,
		"name":            "updated-token",
		"expired_time":    -1,
		"remain_quota":    2000,
		"user_id":         5,
	}
	jsonBody, _ := json.Marshal(tokenBody)
	c.Request.Body = io.NopCloser(bytes.NewReader(jsonBody))

	UpdateToken(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["success"] == true {
		data := resp["data"].(map[string]interface{})
		assert.NotEqual(t, float64(5), data["user_id"])
	}
}