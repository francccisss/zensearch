package utilities

import (
	"fmt"
	"net/url"
)

func GetOrigin(currentUrl string) (string, error) {
	u, err := url.Parse(currentUrl)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	return u.Hostname(), nil
}
