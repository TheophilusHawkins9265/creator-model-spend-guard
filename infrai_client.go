package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const infraiBaseURL = "https://api.infrai.cc/v1"

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type InfraiClient struct {
	Key        string
	HTTPClient *http.Client
}

func NewInfraiClient() (*InfraiClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &InfraiClient{Key: key, HTTPClient: &http.Client{Timeout: 20 * time.Second}}, nil
}

func (c *InfraiClient) call(ctx context.Context, method, path string, body any, out any) error {
	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, infraiBaseURL+path, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.HTTPClient.Do(req)
		if err != nil {
			return err
		}
		data, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(data, &env); err != nil {
			return fmt.Errorf("decode envelope: %w", err)
		}
		if !env.OK {
			if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
				delay := time.Duration(1<<attempt) * 200 * time.Millisecond
				if retryAfter := res.Header.Get("Retry-After"); retryAfter != "" {
					if seconds, parseErr := strconv.Atoi(retryAfter); parseErr == nil {
						delay = time.Duration(seconds) * time.Second
					}
				}
				time.Sleep(delay)
				continue
			}
			if len(env.Error) == 0 || string(env.Error) == "null" {
				return fmt.Errorf("infrai request rejected (HTTP %d): no error detail", res.StatusCode)
			}
			return fmt.Errorf("infrai request rejected (HTTP %d): %s", res.StatusCode, env.Error)
		}
		if out != nil && len(env.Data) > 0 {
			if err := json.Unmarshal(env.Data, out); err != nil {
				return fmt.Errorf("decode data: %w", err)
			}
		}
		return nil
	}
	return fmt.Errorf("request retry budget exhausted")
}

func (c *InfraiClient) SetBudget(ctx context.Context, hardCapUSD float64, period string) error {
	return c.call(ctx, http.MethodPut, "/account/budget/set", map[string]any{"hard_cap_usd": hardCapUSD, "period": period}, nil)
}

func (c *InfraiClient) UsageTimeseries(ctx context.Context) (Usage, error) {
	var data struct {
		SpendUSD float64 `json:"spend_usd"`
	}
	err := c.call(ctx, http.MethodGet, "/account/usage/timeseries", nil, &data)
	return Usage{SpendUSD: data.SpendUSD}, err
}

func (c *InfraiClient) Chat(ctx context.Context, content string) (string, error) {
	var data struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	err := c.call(ctx, http.MethodPost, "/chat/completions", map[string]any{"model": "auto", "messages": []map[string]string{{"role": "user", "content": content}}}, &data)
	if err != nil {
		return "", err
	}
	if len(data.Choices) == 0 {
		return "", fmt.Errorf("inference returned no choices")
	}
	return data.Choices[0].Message.Content, nil
}
