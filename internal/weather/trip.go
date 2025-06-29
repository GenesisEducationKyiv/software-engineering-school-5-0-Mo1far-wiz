package weather

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type WeatherLoggingRoundTripper struct {
	Base    http.RoundTripper
	Logger  Logger
	APIName string
}

func (w *WeatherLoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	base := w.Base
	if base == nil {
		base = http.DefaultTransport
	}

	var reqBody []byte
	if req.Body != nil {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			w.Logger.LogError(fmt.Sprintf("%s – error reading request body: %v", w.APIName, err))
		}
		reqBody = buf
		req.Body = io.NopCloser(bytes.NewBuffer(buf))
	}

	resp, err := base.RoundTrip(req)
	if err != nil {
		w.Logger.LogError(fmt.Sprintf(
			"%s – request to %s failed: %v | request_body=%q",
			w.APIName, req.URL.String(), err, string(reqBody),
		))
		return nil, err
	}

	respBody, readErr := io.ReadAll(resp.Body)
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))
	if readErr != nil {
		w.Logger.LogError(fmt.Sprintf(
			"%s – error reading response body: %v",
			w.APIName, readErr,
		))
		return resp, nil
	}

	logLine := fmt.Sprintf(
		"%s – URL=%s | Status=%d | RequestBody=%q | ResponseBody=%q",
		w.APIName,
		req.URL.String(),
		resp.StatusCode,
		string(reqBody),
		string(respBody),
	)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		w.Logger.LogInfo(logLine)
	} else {
		w.Logger.LogError(logLine)
	}

	return resp, nil
}
