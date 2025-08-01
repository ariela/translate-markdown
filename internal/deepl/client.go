package deepl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	apiURLEndpoint = "https://api-free.deepl.com/v2/translate"
	maxRetries     = 3
	initialBackoff = 1 * time.Second
)

var formalitySupportedLanguages = map[string]bool{
	"DE": true, "FR": true, "IT": true, "ES": true, "NL": true,
	"PL": true, "PT-PT": true, "PT-BR": true, "RU": true, "JA": true,
}

// ClientはDeepL APIとの通信を管理します。
type Client struct {
	apiKey     string
	httpClient *http.Client
	logger     *slog.Logger
}

// NewClientは新しいDeepLクライアントを作成します。
func NewClient(apiKey string, logger *slog.Logger) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger: logger,
	}
}

// TranslateRequestはAPIへのリクエストボディの構造です。
type TranslateRequest struct {
	Text       []string `json:"text"`
	TargetLang string   `json:"target_lang"`
	Formality  string   `json:"formality,omitempty"`
}

// TranslateResponseはAPIからの成功レスポンスボディの構造です。
type TranslateResponse struct {
	Translations []struct {
		Text string `json:"text"`
	} `json:"translations"`
}

// ErrorResponseはAPIからのエラーレスポンスボディの構造です。
type ErrorResponse struct {
	Message string `json:"message"`
}

// TranslateはテキストのスライスをDeepL APIに送信して翻訳します。
func (c *Client) Translate(texts []string, targetLang string) ([]string, error) {
	if len(texts) == 0 {
		return []string{}, nil
	}

	reqBody := TranslateRequest{Text: texts, TargetLang: targetLang}
	if formalitySupportedLanguages[targetLang] {
		reqBody.Formality = "more"
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	var lastErr error
	backoff := initialBackoff

	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("POST", apiURLEndpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "DeepL-Auth-Key "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")

		c.logger.Debug("Sending DeepL API request", "attempt", i+1, "url", apiURLEndpoint, "body", string(jsonData))

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		if resp.StatusCode == http.StatusOK {
			var translateResp TranslateResponse
			if err := json.NewDecoder(resp.Body).Decode(&translateResp); err != nil {
				resp.Body.Close()
				return nil, fmt.Errorf("failed to decode successful response: %w", err)
			}
			resp.Body.Close()

			var translatedTexts []string
			for _, t := range translateResp.Translations {
				translatedTexts = append(translatedTexts, t.Text)
			}
			return translatedTexts, nil
		}

		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to read error response body: %w", readErr)
		}
		resp.Body.Close()
		c.logger.Debug("Received DeepL API error response", "status", resp.Status, "body", string(body))

		var errorResp ErrorResponse
		if json.Unmarshal(body, &errorResp) == nil && errorResp.Message != "" {
			lastErr = fmt.Errorf("API request failed with status %s: %s", resp.Status, errorResp.Message)
		} else {
			lastErr = fmt.Errorf("API request failed with status %s", resp.Status)
		}

		if resp.StatusCode == 456 { // Quota exceeded
			return nil, lastErr
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			c.logger.Warn("Rate limit or server error. Retrying...", "backoff", backoff)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		return nil, lastErr
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
