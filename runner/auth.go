package runner

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type AuthJson struct {
	Expires_at string `json:"expires_at"`
	Token      string `json:"token"`
}

func getToken() (string, error) {
	clientId := os.Getenv("clientId")
	apiKey := os.Getenv("apiKey")

	requestURL := fmt.Sprintf("%s/api/v1/authentication/login", os.Getenv("apiUrl"))
	req, err := http.NewRequest("POST", requestURL, nil)

	req.Header = http.Header{"Content-Type": {"application/json"}, "x-client-id": {clientId}, "x-api-key": {apiKey}}
	if err != nil {
		return "", fmt.Errorf("could not create http request: %s", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making http request: %s", err)
	}

	defer res.Body.Close()

	var authJson AuthJson
	if err := json.NewDecoder(res.Body).Decode(&authJson); err != nil {
		return "", fmt.Errorf("%s", err)
	}

	return authJson.Token, nil
}
