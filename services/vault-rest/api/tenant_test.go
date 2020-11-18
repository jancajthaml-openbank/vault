package api

import (
    "fmt"
    "strings"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/labstack/echo/v4"
    "github.com/stretchr/testify/assert"

    "github.com/jancajthaml-openbank/vault-rest/system"
)

type mockSystemControl struct {
    system.Control
    units []string
    circuitBreakEnableUnit bool
    circuitBreakDisableUnit bool
}

func (sys mockSystemControl) ListUnits(prefix string) ([]string, error) {
    var result = make([]string, 0)
    for _, unit := range sys.units {
        if !strings.HasPrefix(unit, prefix) {
            continue
        }
        result = append(result, strings.TrimSuffix(strings.TrimPrefix(unit, prefix), ".service"))
    }
    return result, nil
}

func (sys mockSystemControl) GetUnitsProperties(name string) (map[string]system.UnitStatus, error) {
    return make(map[string]system.UnitStatus), nil
}

func (sys mockSystemControl) DisableUnit(name string) error {
    if sys.circuitBreakDisableUnit {
        return fmt.Errorf("disable unit circuit break")
    }
    return nil
}

func (sys mockSystemControl) EnableUnit(name string) error {
    if sys.circuitBreakEnableUnit {
        return fmt.Errorf("enable unit circuit break")
    }
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

    t.Log("enable unit fails")
    {
        mockControl := new(mockSystemControl)
        mockControl.circuitBreakEnableUnit = true

        router := echo.New()
        router.POST("/tenant/:tenant", CreateTenant(mockControl))

        req := httptest.NewRequest(http.MethodPost, "/tenant/x", nil)
        rec := httptest.NewRecorder()
        router.ServeHTTP(rec, req)

        assert.Equal(t, http.StatusInternalServerError, rec.Code)
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

    t.Log("missing tenant")
    {
        mockControl := new(mockSystemControl)

        router := echo.New()
        router.DELETE("/tenant/:tenant", DeleteTenant(mockControl))

        req := httptest.NewRequest(http.MethodDelete, "/tenant/ ", nil)
        rec := httptest.NewRecorder()
        router.ServeHTTP(rec, req)

        assert.Equal(t, http.StatusNotFound, rec.Code)
    }

    t.Log("disable unit fails")
    {
        mockControl := new(mockSystemControl)
        mockControl.circuitBreakDisableUnit = true

        router := echo.New()
        router.DELETE("/tenant/:tenant", DeleteTenant(mockControl))

        req := httptest.NewRequest(http.MethodDelete, "/tenant/x", nil)
        rec := httptest.NewRecorder()
        router.ServeHTTP(rec, req)

        assert.Equal(t, http.StatusInternalServerError, rec.Code)
    }
}

func TestGetTenants(t *testing.T) {
    t.Log("happy path")
    {
        mockControl := new(mockSystemControl)
        mockControl.units = append(mockControl.units, "vault-unit@a.service")
        mockControl.units = append(mockControl.units, "vault-unit@b.service")

        router := echo.New()
        router.GET("/tenant", ListTenants(mockControl))

        req := httptest.NewRequest(http.MethodGet, "/tenant", nil)
        rec := httptest.NewRecorder()
        router.ServeHTTP(rec, req)

        assert.Equal(t, http.StatusOK, rec.Code)
        assert.Equal(t, "a\nb", rec.Body.String())
    }
}
