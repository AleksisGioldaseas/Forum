package config

// Contains configuration for external service signup/login

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	customErr "forum/common/custom_errs"
)

type OAuthProvider struct {
	ClientID        string   `json:"client_id"`
	ClientSecret    string   `json:"client_secret"`
	Scopes          []string `json:"scopes"`
	AuthURL         string   `json:"auth_url"`
	TokenURL        string   `json:"token_url"`
	BaseRedirectURI string   `json:"base_redirect_uri"`
}

type OAuthConfig struct {
	Google OAuthProvider `json:"google"`
	Github OAuthProvider `json:"github"`
}

var (
	GoogleOAuth *OAuthProvider
	GithubOAuth *OAuthProvider
)

func InitOAuthConfig(oauthConfig *OAuthConfig, useHTTPS bool) {
	GoogleOAuth = &OAuthProvider{
		ClientID:        oauthConfig.Google.ClientID,
		ClientSecret:    oauthConfig.Google.ClientSecret,
		Scopes:          oauthConfig.Google.Scopes,
		AuthURL:         oauthConfig.Google.AuthURL,
		TokenURL:        oauthConfig.Google.TokenURL,
		BaseRedirectURI: oauthConfig.Google.BaseRedirectURI,
	}

	GithubOAuth = &OAuthProvider{
		ClientID:        oauthConfig.Github.ClientID,
		ClientSecret:    oauthConfig.Github.ClientSecret,
		Scopes:          oauthConfig.Github.Scopes,
		AuthURL:         oauthConfig.Github.AuthURL,
		TokenURL:        oauthConfig.Github.TokenURL,
		BaseRedirectURI: oauthConfig.Github.BaseRedirectURI,
	}
}

func (c *OAuthProvider) ExchangeCodeForToken(code string) (string, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)
	data.Set("redirect_uri", c.getRedirectURL(true))
	data.Set("grant_type", "authorization_code")

	r, err := http.NewRequest("POST", c.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(r)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return "", fmt.Errorf("oauth: cannot fetch token %d\nresponse: %s", response.StatusCode, body)
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	if err := json.NewDecoder(response.Body).Decode(&tokenResponse); err != nil {
		return "", err
	}

	if tokenResponse.AccessToken == "" || tokenResponse.TokenType == "" {
		return "", customErr.ErrOAuthMissingToken
	}

	return tokenResponse.AccessToken, nil
}

func (c *OAuthProvider) GetAuthURL(state string) string {
	u, _ := url.Parse(c.AuthURL)
	q := u.Query()
	q.Set("client_id", c.ClientID)
	q.Set("redirect_uri", c.getRedirectURL(true))
	q.Set("response_type", "code")
	q.Set("state", state)
	q.Set("scope", strings.Join(c.Scopes, " "))
	u.RawQuery = q.Encode()
	return u.String()
}

func (c *OAuthProvider) getRedirectURL(useHTTPS bool) string {
	protocol := "http"
	if useHTTPS {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s", protocol, c.BaseRedirectURI)
}
