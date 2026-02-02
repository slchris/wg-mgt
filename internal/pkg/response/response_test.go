package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]string{"message": "hello"}

	JSON(rec, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var result map[string]string
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "hello", result["message"])
}

func TestSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	Success(rec, data)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result Response
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotNil(t, result.Data)
}

func TestCreated(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]string{"id": "123"}

	Created(rec, data)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var result Response
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestNoContent(t *testing.T) {
	rec := httptest.NewRecorder()

	NoContent(rec)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestError(t *testing.T) {
	rec := httptest.NewRecorder()

	Error(rec, http.StatusBadRequest, "bad request")

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var result Response
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "bad request", result.Error)
}

func TestBadRequest(t *testing.T) {
	rec := httptest.NewRecorder()

	BadRequest(rec, "invalid input")

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var result Response
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "invalid input", result.Error)
}

func TestUnauthorized(t *testing.T) {
	rec := httptest.NewRecorder()

	Unauthorized(rec, "not authenticated")

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var result Response
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "not authenticated", result.Error)
}

func TestForbidden(t *testing.T) {
	rec := httptest.NewRecorder()

	Forbidden(rec, "access denied")

	assert.Equal(t, http.StatusForbidden, rec.Code)

	var result Response
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "access denied", result.Error)
}

func TestNotFound(t *testing.T) {
	rec := httptest.NewRecorder()

	NotFound(rec, "resource not found")

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var result Response
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "resource not found", result.Error)
}

func TestInternalError(t *testing.T) {
	rec := httptest.NewRecorder()

	InternalError(rec, "internal error")

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var result Response
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "internal error", result.Error)
}
