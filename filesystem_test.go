package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"

	"github.com/pyrra-dev/pyrra/slo"
)

type TestLogger struct {
	t *testing.T
}

func NewTestLogger(t *testing.T) *TestLogger {
	return &TestLogger{t: t}
}

func (tl *TestLogger) Log(keyvals ...interface{}) error {
	for i := 0; i < len(keyvals); i += 2 {
		tl.t.Logf("%v=%v", keyvals[i], keyvals[i+1])
	}
	return nil
}

func TestListSpecsHandler(t *testing.T) {
	testLogger := NewTestLogger(t)

	req, err := http.NewRequest("GET", "/specs/list", nil)
	require.NoError(t, err)

	specsDir, err := os.MkdirTemp("", "pyrra-test1")
	require.NoError(t, err)
	defer os.RemoveAll(specsDir)

	file, err := os.Create(specsDir + "/foobar.yaml")
	require.NoError(t, err)
	defer file.Close()

	prometheusDir, err := os.MkdirTemp("", "pyrra-test2")
	require.NoError(t, err)
	defer os.RemoveAll(prometheusDir)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(listSpecsHandler(testLogger, specsDir, prometheusDir))
	handler.ServeHTTP(rr, req)

	require.Equal(t, rr.Code, http.StatusOK, "Should return OK")

	var payload SpecsList
	json.Unmarshal(rr.Body.Bytes(), &payload)
	specsAvailable := payload.SpecsAvailable
	rulesGenerated := payload.RulesGenerated

	require.Contains(t, specsAvailable, "foobar.yaml", "Unexpected value for the SpecsAvailable field in response payload")
	require.Empty(t, rulesGenerated, "The RulesGenerated field in response payload should be empty")
}

func TestCreateSpecHandlerOkSpec(t *testing.T) {
	testLogger := NewTestLogger(t)
	reload := make(chan struct{}, 16)
	fileContent := `---
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: foo
  namespace: foobar
  labels:
    component: foo
spec:
  description: foo
  target: '99.9'
  window: 30d
  indicator:
    ratio:
      errors:
        metric: http_requests_total{name="foobarrr",status=~"5.+"}
      total:
        metric: http_requests_total{name="foobarrr"}`

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	fileWriter, err := writer.CreateFormFile("spec", "foo-spec.yaml")
	require.NoError(t, err)

	_, err = fileWriter.Write([]byte(fileContent))
	require.NoError(t, err)
	writer.Close()

	req, err := http.NewRequest("POST", "/specs/create", &body)
	require.NoError(t, err)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	tmpDir1, err := os.MkdirTemp("", "pyrra-test1")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir1)

	tmpDir2, err := os.MkdirTemp("", "pyrra-test2")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir2)

	handler := http.HandlerFunc(createSpecHandler(testLogger, tmpDir1, tmpDir2, reload, false))
	handler.ServeHTTP(rr, req)

	require.Equal(t, rr.Code, http.StatusOK, "Should return OK")
}

func TestCreateSpecHandlerBadSpec(t *testing.T) {
	testLogger := NewTestLogger(t)
	reload := make(chan struct{}, 16)
	fileContent := "garbage"

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	fileWriter, err := writer.CreateFormFile("spec", "test.yaml")
	if err != nil {
		t.Fatal(err)
	}

	_, err = fileWriter.Write([]byte(fileContent))
	if err != nil {
		t.Fatal(err)
	}
	writer.Close()

	req := httptest.NewRequest("POST", "/specs/create", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(createSpecHandler(testLogger, "/tmp/", "/tmp/", reload, false))
	handler.ServeHTTP(rr, req)

	require.Equal(t, rr.Code, http.StatusBadRequest, "Should return BadRequest")
}

func TestCreateSpecHandlerInvalidMethod(t *testing.T) {
	logger := NewTestLogger(t)
	reload := make(chan struct{}, 16)

	req, err := http.NewRequest("PUT", "/specs/create", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(createSpecHandler(logger, "/tmp/pyrra/1", "/tmp/pyrra/2", reload, false))
	handler.ServeHTTP(rr, req)

	require.Equal(t, rr.Code, http.StatusMethodNotAllowed, "Should return MethodNotAllowed")
}

func TestCreateSpecHandlerMissingParameter(t *testing.T) {
	logger := NewTestLogger(t)
	reload := make(chan struct{}, 16)

	req, err := http.NewRequest("POST", "/specs/create", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(createSpecHandler(logger, "/tmp/pyrra/1", "/tmp/pyrra/2", reload, false))
	handler.ServeHTTP(rr, req)

	require.Equal(t, rr.Code, http.StatusBadRequest, "Should return BadRequest")
}

func TestRemoveSpecHandlerInvalidMethod(t *testing.T) {
	logger := NewTestLogger(t)
	reload := make(chan struct{}, 16)

	req, err := http.NewRequest("GET", "/specs/remove", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(removeSpecHandler(logger, "/tmp/pyrra/1", "/tmp/pyrra/2", reload))
	handler.ServeHTTP(rr, req)

	require.Equal(t, rr.Code, http.StatusMethodNotAllowed, "Should return MethodNotAllowed")
}

func TestRemoveSpecHandlerMissingParameter(t *testing.T) {
	logger := NewTestLogger(t)
	reload := make(chan struct{}, 16)

	req, err := http.NewRequest("DELETE", "/specs/remove", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(removeSpecHandler(logger, "/tmp/pyrra/1", "/tmp/pyrra/2", reload))
	handler.ServeHTTP(rr, req)

	require.Equal(t, rr.Code, http.StatusBadRequest, "Should return BadRequest")

	expectedBody := "Missing 'f' parameter in the query"

	require.Equal(t, rr.Code, http.StatusBadRequest, "Should return BadRequest")
	require.Equal(t, strings.TrimSpace(rr.Body.String()), expectedBody, "Should return expected body")
}

func TestMatchObjectives(t *testing.T) {
	obj1 := slo.Objective{Labels: labels.FromStrings("foo", "bar")}
	obj2 := slo.Objective{Labels: labels.FromStrings("foo", "bar", "ying", "yang")}
	obj3 := slo.Objective{Labels: labels.FromStrings("foo", "bar", "yes", "no")}
	obj4 := slo.Objective{Labels: labels.FromStrings("foo", "baz")}

	objectives := Objectives{objectives: map[string]slo.Objective{}}
	objectives.Set(obj1)
	objectives.Set(obj2)
	objectives.Set(obj3)
	objectives.Set(obj4)

	matches := objectives.Match([]*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "foo", "foo"),
	})
	require.Nil(t, matches)

	matches = objectives.Match([]*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "foo", "bar"),
	})
	require.Len(t, matches, 3)
	require.Contains(t, matches, obj1)
	require.Contains(t, matches, obj2)
	require.Contains(t, matches, obj3)

	matches = objectives.Match([]*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "foo", "baz"),
	})
	require.Len(t, matches, 1)
	require.Contains(t, matches, obj4)

	matches = objectives.Match([]*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "foo", "bar"),
		labels.MustNewMatcher(labels.MatchEqual, "ying", "yang"),
	})
	require.Len(t, matches, 1)
	require.Contains(t, matches, obj2)

	matches = objectives.Match([]*labels.Matcher{
		labels.MustNewMatcher(labels.MatchRegexp, "foo", "ba."),
	})
	require.Len(t, matches, 4)
	require.Contains(t, matches, obj1)
	require.Contains(t, matches, obj2)
	require.Contains(t, matches, obj3)
	require.Contains(t, matches, obj4)
}
