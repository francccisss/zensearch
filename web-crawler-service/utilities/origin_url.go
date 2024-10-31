package utilities

import (
	"net/url"
)

func GetHostname(currentUrl string) (string, string, error) {
	u, err := url.Parse(currentUrl)
	if err != nil {
		return "", "", err
	}
	return u.Hostname(), u.Path, nil
}
