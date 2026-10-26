package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/engigu/baihu-panel/internal/controllers"
	"github.com/engigu/baihu-panel/internal/services"
	"github.com/gin-gonic/gin"
)

func TestRouter_NoRoute_API_JSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c := &Controllers{
		File: controllers.NewFileController(t.TempDir()),
	}

	r := Setup(c)

	// 1. 测试 /api/v1 未匹配路径返回 404 JSON
	reqAPI, _ := http.NewRequest(http.MethodGet, "/api/v1/non_existent_endpoint", nil)
	wAPI := httptest.NewRecorder()
	r.ServeHTTP(wAPI, reqAPI)

	if wAPI.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 for /api/v1 404, got %d", wAPI.Code)
	}

	var jsonAPI struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(wAPI.Body.Bytes(), &jsonAPI); err != nil {
		t.Fatalf("Expected JSON response for /api/v1 404, got error: %v, body: %s", err, wAPI.Body.String())
	}
	if jsonAPI.Code != 404 {
		t.Fatalf("Expected code 404 in body, got %d", jsonAPI.Code)
	}

	// 2. 测试 /open2api/v1 未匹配路径返回 404 JSON 而非 SPA 200 HTML
	reqOpenAPI, _ := http.NewRequest(http.MethodPost, "/open2api/v1/non_existent_endpoint", nil)
	wOpenAPI := httptest.NewRecorder()
	r.ServeHTTP(wOpenAPI, reqOpenAPI)

	if wOpenAPI.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 for /open2api/v1 404, got %d", wOpenAPI.Code)
	}

	var jsonOpenAPI struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(wOpenAPI.Body.Bytes(), &jsonOpenAPI); err != nil {
		t.Fatalf("Expected JSON response for /open2api/v1 404, got error: %v, body: %s", err, wOpenAPI.Body.String())
	}
	if jsonOpenAPI.Code != 404 {
		t.Fatalf("Expected code 404 in body, got %d", jsonOpenAPI.Code)
	}
}

func TestRouter_OpenAPI_FileRoutes_Registered(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c := &Controllers{
		File:   controllers.NewFileController(t.TempDir()),
		Script: controllers.NewScriptController(services.NewScriptService()),
	}

	r := Setup(c)
	routes := r.Routes()

	expectedRoutes := map[string]string{
		"GET /open2api/v1/files/tree":    "",
		"GET /open2api/v1/files/search":  "",
		"GET /open2api/v1/files/content": "",
		"POST /open2api/v1/files/content": "",
		"POST /open2api/v1/files/create":  "",
		"POST /open2api/v1/files/delete":  "",
	}

	for _, route := range routes {
		key := route.Method + " " + route.Path
		delete(expectedRoutes, key)
	}

	if len(expectedRoutes) > 0 {
		t.Fatalf("Missing expected OpenAPI file routes: %+v", expectedRoutes)
	}
}
