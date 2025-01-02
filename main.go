package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/rs/zerolog/log"
)

type GetTokenResDto struct {
	AccessToken      string `json:"access_token" validate:"required"`
	ExpiresIn        int    `json:"expires_in" validate:"required"` //expires in seconds
	RefreshToken     string `json:"refresh_token" validate:"required"`
	RefreshExpiresIn int    `json:"refresh_expires_in" validate:"required"` //expires in seconds
	TokenType        string `json:"token_type" validate:"required"`
	NotBeforePolicy  int    `json:"not-before-policy" validate:"required"`
	SessionState     string `json:"session_state" validate:"required"`
	Scope            string `json:"scope" validate:"required"`
}

var (
	bboneTokenKey   string                                   = "bbone_token"
	bboneTokenCache *ttlcache.Cache[string, *GetTokenResDto] = ttlcache.New[string, *GetTokenResDto]()
)

func GetToken(ctx context.Context) (*GetTokenResDto, error) {
	// _, span := tracer.StartSpan(ctx, "ibl_billing.bb_auth.GetToken", nil)
	// defer span.End()

	if bboneTokenCache.Has(bboneTokenKey) {
		bboneToken := bboneTokenCache.Get(bboneTokenKey).Value()
		return bboneToken, nil
	}

	// path := "/realms/BBG/protocol/openid-connect/token"
	// headres := map[string]string{}
	// headres["Content-Type"] = "application/x-www-form-urlencoded"

	// Create URL values for form data
	data := url.Values{}
	data.Set("client_id", "i")
	data.Set("username", "n")
	data.Set("password", "E")
	data.Set("grant_type", "p")
	data.Set("client_secret", "H")
	// Convert to io.Reader
	reader := strings.NewReader(data.Encode())

	// Create new request
	// todo: refactor this api call
	// url := fmt.Sprintf("%s/protocol/openid-connect/token", m.authBaseUrl)
	url := "https://example/realms/BBG/protocol/openid-connect/token"
	req, err := http.NewRequestWithContext(ctx, "POST", url, reader)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return nil, err
	}

	// Add headers
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Create HTTP client and send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return nil, err
	}

	// Parse response
	result := &GetTokenResDto{}
	if err := json.Unmarshal(body, result); err != nil {
		log.Error().Err(err).Msg("failed to parse token response")
		return nil, fmt.Errorf("parse token response: %w", err)
	}

	expInMinute := (result.ExpiresIn * 60) - 1
	expDurationInMinute := time.Duration(expInMinute) * time.Minute
	bboneTokenCache.Set(bboneTokenKey, result, expDurationInMinute)

	return result, nil
}

func main() {
	for i := 0; i < 10; i++ {
		GetToken(context.Background())
		if i == 5 {
			// Reset cache
			bboneTokenCache = ttlcache.New[string, *GetTokenResDto]()
		}
	}
}
