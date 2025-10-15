package responsex

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetaMiddlewareSetsRequestIDBeforeHandler(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.Use(MetaMiddleware("v1.2.3"))
	engine.GET("/meta", func(c *gin.Context) {
		OK(c, gin.H{"status": "ok"}, nil)
	})

	req, err := http.NewRequest(http.MethodGet, "/meta", nil)
	require.NoError(t, err)
	const expectedRID = "test-request-id"
	req.Header.Set("X-Request-Id", expectedRID)

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	assert.Equal(t, expectedRID, rec.Header().Get("X-Request-Id"))

	var resp Envelope[map[string]any]
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.True(t, resp.Ok)
	require.NotNil(t, resp.Meta)
	assert.Equal(t, expectedRID, resp.Meta.RequestID)
	assert.Equal(t, "v1.2.3", resp.Meta.Server)
}
