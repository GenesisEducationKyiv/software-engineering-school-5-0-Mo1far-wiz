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
	URL     string
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
			w.Logger.Error(fmt.Sprintf("%s – error reading request body: %v", w.APIName, err))
		}
		reqBody = buf
		req.Body = io.NopCloser(bytes.NewBuffer(buf))
	}

	resp, err := base.RoundTrip(req)
	if err != nil {
		w.Logger.Error(fmt.Sprintf(
			"%s – request to %s failed: %v | request_body=%q",
			w.APIName, req.URL.String(), err, string(reqBody),
		))
		return nil, err
	}

	respBody, readErr := io.ReadAll(resp.Body)
	if cerr := resp.Body.Close(); cerr != nil {
		w.Logger.Error(fmt.Sprintf(
			"%s – error closing response body: %v",
			w.APIName, cerr,
		))
	}

	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

	if readErr != nil {
		w.Logger.Error(fmt.Sprintf(
			"%s – error reading response body: %v",
			w.APIName, readErr,
		))
		return resp, nil
	}

	logLine := fmt.Sprintf(
		"%s\n\tURL:\t%s \n\tStatus:\t%d \n\tRequestBody:\t%q \n\tResponseBody:\t%q",
		w.APIName,
		w.URL,
		resp.StatusCode,
		string(reqBody),
		string(respBody),
	)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		w.Logger.Info(logLine)
	} else {
		w.Logger.Error(logLine)
	}

	return resp, nil
}
