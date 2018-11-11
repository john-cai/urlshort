package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/john-cai/urlshort/database"
	"github.com/john-cai/urlshort/models"
)

func setUpTestServer(t *testing.T) *Server {
	db, err := database.NewTestDB()
	require.NoError(t, err)
	s := &Server{
		database: db,
		Router:   mux.NewRouter(),
	}
	s.configureRoutes()
	return s
}

func TestShortenValidation(t *testing.T) {
	s := setUpTestServer(t)
	tc := []struct {
		input        models.URLRequest
		responseCode int
	}{
		{input: models.URLRequest{Url: "https://www.mywebpage.com"}, responseCode: http.StatusOK},
		{input: models.URLRequest{Url: "http://www.range.io"}, responseCode: http.StatusOK},
		{input: models.URLRequest{Url: ""}, responseCode: http.StatusBadRequest},
	}

	for _, testCase := range tc {
		rec := httptest.NewRecorder()
		var b bytes.Buffer
		json.NewEncoder(&b).Encode(&testCase.input)
		req := httptest.NewRequest(http.MethodPost, "/shorten", &b)
		s.Shorten(rec, req)
		assert.Equal(t, testCase.responseCode, rec.Result().StatusCode)

	}
}

func TestShortenHappyPath(t *testing.T) {
	s := setUpTestServer(t)

	input := models.URLRequest{
		Url: fmt.Sprintf("http://%s", uuid.New()),
	}
	rec := httptest.NewRecorder()
	var b bytes.Buffer
	json.NewEncoder(&b).Encode(&input)
	req := httptest.NewRequest(http.MethodPost, "/shorten", &b)
	s.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Result().StatusCode)
	var url models.URL
	require.NoError(t, json.NewDecoder(rec.Result().Body).Decode(&url))

	//check if the redirect happens

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, url.Short, nil)
	s.ServeHTTP(rec, req)
	require.Equal(t, http.StatusSeeOther, rec.Result().StatusCode)
	assert.Equal(t, input.Url, rec.Header()["Location"][0])
}

func TestShortenCustom(t *testing.T) {
	s := setUpTestServer(t)

	url := fmt.Sprintf("http://%s", uuid.New())
	custom := "range"
	input := models.URLRequest{
		Url:    url,
		Custom: &custom,
	}
	rec := httptest.NewRecorder()
	var b bytes.Buffer
	json.NewEncoder(&b).Encode(&input)
	req := httptest.NewRequest(http.MethodPost, "/shorten", &b)
	s.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Result().StatusCode)
	var u models.URL
	require.NoError(t, json.NewDecoder(rec.Result().Body).Decode(&u))
	assert.Equal(t, fmt.Sprintf("/links/%s", custom), u.Short)

	//check if the redirect happens
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, u.Short, nil)
	s.ServeHTTP(rec, req)
	require.Equal(t, http.StatusSeeOther, rec.Result().StatusCode)
	assert.Equal(t, input.Url, rec.Header()["Location"][0])

	//trying to create another link with the same custom link will fail with a 400
	input = models.URLRequest{
		Url:    fmt.Sprintf("http://%s", uuid.New()),
		Custom: &custom,
	}

	rec = httptest.NewRecorder()
	b = bytes.Buffer{}
	json.NewEncoder(&b).Encode(&input)
	req = httptest.NewRequest(http.MethodPost, "/shorten", &b)
	s.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)

	// trying to create another link with a bad custom link will fail with a 400

	custom = "full of % &* characters"
	input = models.URLRequest{
		Url:    fmt.Sprintf("http://%s", uuid.New()),
		Custom: &custom,
	}

	rec = httptest.NewRecorder()
	b = bytes.Buffer{}
	json.NewEncoder(&b).Encode(&input)
	req = httptest.NewRequest(http.MethodPost, "/shorten", &b)
	s.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
}
