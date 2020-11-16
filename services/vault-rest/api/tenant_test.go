package api

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/labstack/echo/v4"
    "github.com/stretchr/testify/assert"

    "github.com/jancajthaml-openbank/vault-rest/system"
)

type mockSystemControl struct {
    system.Control
}

func (sys mockSystemControl) ListUnits(name string) ([]string, error) {
    return nil, nil
}

func (sys mockSystemControl) GetUnitsProperties(name string) (map[string]system.UnitStatus, error) {
    return make(map[string]system.UnitStatus), nil
}

func (sys mockSystemControl) DisableUnit(name string) error {
    return nil
}

func (sys mockSystemControl) EnableUnit(name string) error {
    return nil
}

func TestCreateTenant(t *testing.T) {
    t.Log("happy path")
    {
        mockControl := new(mockSystemControl)

        router := echo.New()
        router.POST("/tenant/:tenant", CreateTenant(mockControl))

        req := httptest.NewRequest(http.MethodPost, "/tenant/x", nil)
        rec := httptest.NewRecorder()
        router.ServeHTTP(rec, req)

        assert.Equal(t, http.StatusOK, rec.Code)
        assert.Equal(t, "", rec.Body.String())
    }

    t.Log("missing tenant")
    {
        mockControl := new(mockSystemControl)

        router := echo.New()
        router.POST("/tenant/:tenant", CreateTenant(mockControl))

        req := httptest.NewRequest(http.MethodPost, "/tenant/ ", nil)
        rec := httptest.NewRecorder()
        router.ServeHTTP(rec, req)

        assert.Equal(t, http.StatusNotFound, rec.Code)
    }
}

func TestGetTenants(t *testing.T) {
    t.Log("happy path")
    {
        mockControl := new(mockSystemControl)

        router := echo.New()
        router.GET("/tenant", ListTenants(mockControl))

        req := httptest.NewRequest(http.MethodGet, "/tenant", nil)
        rec := httptest.NewRecorder()
        router.ServeHTTP(rec, req)

        assert.Equal(t, http.StatusOK, rec.Code)
        assert.Equal(t, "", rec.Body.String())
    }
}

func TestDeleteTenant(t *testing.T) {
    t.Log("happy path")
    {
        mockControl := new(mockSystemControl)

        router := echo.New()
        router.DELETE("/tenant/:tenant", DeleteTenant(mockControl))

        req := httptest.NewRequest(http.MethodDelete, "/tenant/x", nil)
        rec := httptest.NewRecorder()
        router.ServeHTTP(rec, req)

        assert.Equal(t, http.StatusOK, rec.Code)
        assert.Equal(t, "", rec.Body.String())
    }
}
