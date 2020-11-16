package api

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/labstack/echo/v4"
    "github.com/stretchr/testify/assert"

    "github.com/jancajthaml-openbank/vault-rest/system"
)

type mockMonitor struct {
    system.MemoryMonitor
    system.DiskMonitor
}

func (monitoy mockMonitor) IsHealthy() bool {
    return true
}

func (monitoy mockMonitor) GetFreeMemory() uint64 {
    return uint64(0)
}

func (monitoy mockMonitor) GetUsedMemory() uint64 {
    return uint64(0)
}

func (monitoy mockMonitor) GetFreeDiskSpace() uint64 {
    return uint64(0)
}

func (monitoy mockMonitor) GetUsedDiskSpace() uint64 {
    return uint64(0)
}

func TestHealthCheckHandler(t *testing.T) {
    t.Log("HEAD /health")
    {
        monitor := new(mockMonitor)

        router := echo.New()
        router.HEAD("/health", HealtCheckPing(monitor, monitor))

        req := httptest.NewRequest(http.MethodHead, "/health", nil)
        rec := httptest.NewRecorder()
        router.ServeHTTP(rec, req)

        assert.Equal(t, http.StatusOK, rec.Code)
        assert.Empty(t, rec.Body.String())
    }

    t.Log("GET /health")
    {
        monitor := new(mockMonitor)

        router := echo.New()
        router.GET("/health", HealtCheck(monitor, monitor))

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
