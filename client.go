package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const (
	userAgent       = "KilimcininKorOglu/kamusm-go/" + version
	maxResponseSize = 1 << 16 // 64KB
)

// setCommonHeaders sets the common HTTP headers for all KamuSM requests.
func setCommonHeaders(req *http.Request, identity string) {
	req.Header.Set("Content-Type", "application/timestamp-query")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("identity", identity)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Accept", "text/html, image/gif, image/jpeg, */*; q=0.2")
	req.Header.Set("Connection", "keep-alive")
}

// sendTimestampRequest sends a timestamp request to the KamuSM server.
func sendTimestampRequest(host, identity string, der []byte) (int, []byte, error) {
	req, err := http.NewRequest("POST", host, bytes.NewReader(der))
	if err != nil {
		return 0, nil, fmt.Errorf("istek oluşturulamadı: %w", err)
	}

	setCommonHeaders(req, identity)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("istek gönderilemedi: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return 0, nil, fmt.Errorf("yanıt gövdesi okunamadı: %w", err)
	}

	return resp.StatusCode, body, nil
}

// sendCreditRequest sends a credit balance check request to the KamuSM server.
func sendCreditRequest(host, identity string, customerID uint32, timestamp uint64) (int, string, []byte, error) {
	req, err := http.NewRequest("POST", host, nil)
	if err != nil {
		return 0, "", nil, fmt.Errorf("istek oluşturulamadı: %w", err)
	}

	setCommonHeaders(req, identity)
	req.Header.Set("credit_req", strconv.FormatUint(uint64(customerID), 10))
	req.Header.Set("credit_req_time", strconv.FormatUint(timestamp, 10))
	req.ContentLength = 0

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", nil, fmt.Errorf("bakiye kontrolü isteği gönderilemedi: %w", err)
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return 0, "", nil, fmt.Errorf("yanıt gövdesi okunamadı: %w", err)
	}

	return resp.StatusCode, contentType, body, nil
}
