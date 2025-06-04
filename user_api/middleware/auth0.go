package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

var (
	auth0DomainAdmin   = os.Getenv("AUTH0_DOMAIN")
	clientID           = os.Getenv("AUTH0_MGMT_CLIENT_ID")
	clientSecret       = os.Getenv("AUTH0_MGMT_CLIENT_SECRET")
	auth0Audience      = fmt.Sprintf("https://%s/api/v2/", auth0DomainAdmin)
	auth0TokenURL      = fmt.Sprintf("https://%s/oauth/token", auth0DomainAdmin)
	auth0DeleteUserURL = fmt.Sprintf("https://%s/api/v2/users/", auth0DomainAdmin)
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func GetManagementToken() (string, error) {
	body := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     clientID,
		"client_secret": clientSecret,
		"audience":      auth0Audience,
	}
	payload, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", auth0TokenURL, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Auth0 token error: %s", resp.Status)
	}

	var result tokenResponse
	json.NewDecoder(resp.Body).Decode(&result)

	return result.AccessToken, nil
}

// DeleteUserFromAuth0 deletes a user by Auth0 ID
func DeleteUserFromAuth0(auth0ID string) error {
	token, err := GetManagementToken()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("DELETE", auth0DeleteUserURL+auth0ID, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete user from Auth0: %s", resp.Status)
	}

	return nil
}
