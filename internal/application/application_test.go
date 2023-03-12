package application

import (
	"bytes"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/1g0rbm/sysmonitor/internal/config"
	"github.com/1g0rbm/sysmonitor/internal/metric"
	"github.com/1g0rbm/sysmonitor/internal/storage"
)

const key = "qwerty"

func Test_updateJsonHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		content     string
	}
	fVal := 2.01
	iVal := int64(5)
	tests := []struct {
		name   string
		path   string
		method string
		metric metric.Metrics
		want   want
	}{
		{
			name:   "success update counter metric test",
			path:   "/update/",
			method: http.MethodPost,
			metric: metric.Metrics{
				ID:    "PollCounter",
				MType: metric.CounterType,
				Delta: &iVal,
			},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusOK,
				content:     `{"id":"PollCounter","type":"counter","delta":5}`,
			},
		},
		{
			name:   "success update gauge metric test",
			path:   "/update/",
			method: http.MethodPost,
			metric: metric.Metrics{
				ID:    "Alloc",
				MType: metric.GaugeType,
				Value: &fVal,
			},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusOK,
				content:     `{"id":"Alloc","type":"gauge","value":2.01}`,
			},
		},
		{
			name:   "invalid value update counter metric test",
			path:   "/update/",
			method: http.MethodPost,
			metric: metric.Metrics{
				ID:    "PollCounter",
				MType: metric.CounterType,
			},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusBadRequest,
				content:     "invalid metric value",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp(storage.NewStorage(), config.GetConfigServer())

			ts := httptest.NewServer(app.getRouter())
			defer ts.Close()

			resp, body := testJSONRequest(t, ts, tt.method, tt.path, tt.metric)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.content, body)

			flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
			flag.CommandLine.Init("", flag.ContinueOnError)
		})
	}
}

func Test_updateJsonWithSignHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		content     string
	}
	fVal := 2.01
	iVal := int64(5)
	tests := []struct {
		name    string
		path    string
		method  string
		signKey string
		metric  metric.Metrics
		want    want
	}{
		{
			name:    "success update counter metric test",
			path:    "/update/",
			method:  http.MethodPost,
			signKey: key,
			metric: metric.Metrics{
				ID:    "PollCounter",
				MType: metric.CounterType,
				Delta: &iVal,
			},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusOK,
				content:     `{"id":"PollCounter","type":"counter","delta":5}`,
			},
		},
		{
			name:    "success update gauge metric test",
			path:    "/update/",
			method:  http.MethodPost,
			signKey: key,
			metric: metric.Metrics{
				ID:    "Alloc",
				MType: metric.GaugeType,
				Value: &fVal,
			},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusOK,
				content:     `{"id":"Alloc","type":"gauge","value":2.01}`,
			},
		},
		{
			name:    "invalid value update counter metric test",
			path:    "/update/",
			method:  http.MethodPost,
			signKey: key + "qwerty",
			metric: metric.Metrics{
				ID:    "Alloc",
				MType: metric.GaugeType,
				Value: &fVal,
			},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusBadRequest,
				content:     "wrong sign",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.GetConfigServer()
			cfg.Key = key
			app := NewApp(storage.NewStorage(), cfg)

			ts := httptest.NewServer(app.getRouter())
			defer ts.Close()

			signErr := tt.metric.Sign(tt.signKey)
			require.Nil(t, signErr)

			resp, body := testJSONRequest(t, ts, tt.method, tt.path, tt.metric)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.content, body)

			flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
			flag.CommandLine.Init("", flag.ContinueOnError)
		})
	}
}

func Test_getOneJsonHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		content     string
	}
	tests := []struct {
		name   string
		path   string
		method string
		metric metric.Metrics
		want   want
	}{
		{
			name:   "success get gauge metric test",
			path:   "/value/",
			method: http.MethodPost,
			metric: metric.Metrics{
				ID:    "Alloc",
				MType: metric.GaugeType,
			},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusOK,
				content:     `{"id":"Alloc","type":"gauge","value":390204.95873408683}`,
			},
		},
		{
			name:   "success get counter metric test",
			path:   "/value/",
			method: http.MethodPost,
			metric: metric.Metrics{
				ID:    "PollCounter",
				MType: metric.CounterType,
			},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusOK,
				content:     `{"id":"PollCounter","type":"counter","delta":5}`,
			},
		},
		{
			name:   "get unknown metric test",
			path:   "/value/",
			method: http.MethodPost,
			metric: metric.Metrics{
				ID:    "UnknownCounter",
				MType: metric.CounterType,
			},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusNotFound,
				content:     "metric not found by name 'UnknownCounter'",
			},
		},
	}
	for _, tt := range tests {
		m1, _ := metric.NewMetric("Alloc", metric.GaugeType, "390204.95873408683")
		m2, _ := metric.NewMetric("PollCounter", metric.CounterType, "5")

		t.Run(tt.name, func(t *testing.T) {
			s := storage.NewStorage()
			s.Set(m1)
			s.Set(m2)
			app := NewApp(s, config.GetConfigServer())

			ts := httptest.NewServer(app.getRouter())
			defer ts.Close()

			resp, body := testJSONRequest(t, ts, tt.method, tt.path, tt.metric)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.content, body)

			flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
			flag.CommandLine.Init("", flag.ContinueOnError)
		})
	}
}

func Test_updateHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		content     string
	}
	tests := []struct {
		name   string
		path   string
		method string
		want   want
	}{
		{
			name:   "invalid http method test",
			path:   "/update/counter/PollCounter/175",
			method: http.MethodGet,
			want: want{
				contentType: "",
				statusCode:  http.StatusMethodNotAllowed,
				content:     "",
			},
		},
		{
			name:   "success update gauge metric test",
			path:   "/update/gauge/HeapReleased/2621440.000000",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusOK,
				content:     "",
			},
		},
		{
			name:   "success update counter metric test",
			path:   "/update/counter/PollCounter/5",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusOK,
				content:     "",
			},
		},
		{
			name:   "invalid metric type test",
			path:   "/update/invalid/PollCounter/5",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotImplemented,
				content:     "invalid type invalid\n",
			},
		},
		{
			name:   "invalid path format test",
			path:   "/update/invalid/path/format/PollCounter/5",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
				content:     "404 page not found\n",
			},
		},
		{
			name:   "invalid metric value test",
			path:   "/update/gauge/HeapReleased/aaa",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
				content:     "invalid value\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp(storage.NewStorage(), config.GetConfigServer())

			ts := httptest.NewServer(app.getRouter())
			defer ts.Close()

			resp, body := testRequest(t, ts, tt.method, tt.path)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.content, body)

			flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
			flag.CommandLine.Init("", flag.ContinueOnError)
		})
	}
}

func Test_getOneHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		content     string
	}
	tests := []struct {
		name   string
		path   string
		method string
		want   want
	}{
		{
			name:   "invalid http method test",
			path:   "/value/counter/PollCounter",
			method: http.MethodPost,
			want: want{
				contentType: "",
				statusCode:  http.StatusMethodNotAllowed,
				content:     "",
			},
		},
		{
			name:   "success get gauge metric value test",
			path:   "/value/gauge/HeapReleased",
			method: http.MethodGet,
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusOK,
				content:     "2621440",
			},
		},
		{
			name:   "success get gauge metric value test",
			path:   "/value/gauge/Heap",
			method: http.MethodGet,
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusOK,
				content:     "96969.519",
			},
		},
		{
			name:   "success get counter metric value test",
			path:   "/value/counter/PollCounter",
			method: http.MethodGet,
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusOK,
				content:     "5",
			},
		},
		{
			name:   "get not existed metric value test",
			path:   "/value/gauge/Lookups",
			method: http.MethodGet,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
				content:     "metric not found\n",
			},
		},
		{
			name:   "invalid path format test",
			path:   "/get/invalid/path/format/PollCounter/5",
			method: http.MethodGet,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
				content:     "404 page not found\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp(storage.NewStorage(), config.GetConfigServer())

			ts := httptest.NewServer(app.getRouter())
			defer ts.Close()

			testRequestAndCloseBody(t, ts, "POST", "/update/gauge/HeapReleased/2621440.000000")
			testRequestAndCloseBody(t, ts, "POST", "/update/gauge/Heap/96969.519")
			testRequestAndCloseBody(t, ts, "POST", "/update/counter/PollCounter/5")

			resp, body := testRequest(t, ts, tt.method, tt.path)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.content, body)

			flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
			flag.CommandLine.Init("", flag.ContinueOnError)
		})
	}
}

func Test_getAllHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		content     string
	}
	tests := []struct {
		name   string
		path   string
		method string
		want   want
	}{
		{
			name:   "success get all metrics test",
			path:   "/",
			method: http.MethodGet,
			want: want{
				contentType: "text/html; charset=utf-8",
				statusCode:  http.StatusOK,
				content: `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>All metrics</title>
</head>
<body>
<h1>List of metrics</h1>
<ul>
    
        <li>HeapReleased:2621440</li>
    
        <li>PollCounter:5</li>
    
</ul>
</body>
</html>
`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp(storage.NewStorage(), config.GetConfigServer())

			ts := httptest.NewServer(app.getRouter())
			defer ts.Close()

			testRequestAndCloseBody(t, ts, "POST", "/update/gauge/HeapReleased/2621440.000000")
			testRequestAndCloseBody(t, ts, "POST", "/update/counter/PollCounter/5")

			resp, body := testRequest(t, ts, tt.method, tt.path)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.content, body)

			flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
			flag.CommandLine.Init("", flag.ContinueOnError)
		})
	}
}

func testJSONRequest(
	t *testing.T,
	ts *httptest.Server,
	method,
	path string,
	m metric.Metrics,
) (*http.Response, string) {
	b, _ := m.Encode()
	buf := bytes.NewBuffer(b)

	req, err := http.NewRequest(method, ts.URL+path, buf)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}

func testRequestAndCloseBody(t *testing.T, ts *httptest.Server, method, path string) {
	resp, _ := testRequest(t, ts, method, path)
	resp.Body.Close()
}
