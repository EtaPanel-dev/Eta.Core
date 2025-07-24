package ssl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type ZeroSSLRes struct {
	Success    bool   `json:"success"`
	EabKid     string `json:"eab_kid"`
	EabHmacKey string `json:"eab_hmac_key"`
}

func GetZeroSSLEabCredentials(email string) (*ZeroSSLRes, error) {
	baseURL := "https://api.zerossl.com/acme/eab-credentials-email"
	params := url.Values{}
	params.Add("email", email)
	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequest("POST", requestURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	defer client.CloseIdleConnections()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned non-200 status: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var result ZeroSSLRes
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
