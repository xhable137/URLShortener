package api

import (
	"net/http"
)

// GetRedirect follows redirects and returns the final URL
func GetRedirect(url string) (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusFound {
		return resp.Header.Get("Location"), nil
	}

	return "", nil
}
