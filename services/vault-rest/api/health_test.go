package api

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/labstack/echo/v4"
    "github.com/stretchr/testify/assert"
)

func TestHealthCheckHandler(t *testing.T) {
    t.Log("HEAD /health")
    {
        router := echo.New()

        router.HEAD("/health", HealtCheckPing(nil, nil))

        req := httptest.NewRequest(http.MethodHead, "/health", nil)
        rec := httptest.NewRecorder()
        router.ServeHTTP(rec, req)

        assert.Equal(t, http.StatusOK, rec.Code)
        assert.Empty(t, rec.Body.String())
    }

    t.Log("GET /health")
    {
        router := echo.New()

        router.GET("/health", HealtCheck(nil, nil))

        req := httptest.NewRequest(http.MethodGet, "/health", nil)
        rec := httptest.NewRecorder()
        router.ServeHTTP(rec, req)

        assert.Equal(t, http.StatusOK, rec.Code)
        assert.JSONEq(t, `
            {
                "storage": {
                    "free": 0,
                    "used": 0,
                    "healthy": true
                },
                "memory": {
                    "free": 0,
                    "used": 0,
                    "healthy": true
                }
            }
        `, rec.Body.String())
    }
}
