package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

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
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
				content:     "method not allowed\n",
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
				statusCode:  http.StatusBadRequest,
				content:     "invalid metric type\n",
			},
		},
		{
			name:   "invalid path format test",
			path:   "/update/invalid/path/format/PollCounter/5",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
				content:     "can not extract data from path\n",
			},
		},
		{
			name:   "invalid metric value test",
			path:   "/update/gauge/HeapReleased/aaa",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusInternalServerError,
				content:     "strconv.ParseFloat: parsing \"aaa\": invalid syntax\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			request := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(updateHandler)

			h(w, request)

			result := w.Result()

			require.Equal(t, tt.want.statusCode, result.StatusCode)
			require.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			b, err := io.ReadAll(result.Body)
			if err == nil {
				require.NoError(t, err)
			}

			err = result.Body.Close()
			if err != nil {
				require.NoError(t, err)
			}

			require.Equal(t, tt.want.content, string(b))
		})
	}
}
