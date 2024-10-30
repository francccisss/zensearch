package utilities

import (
	"bufio"
	"net/http"
	"strings"
)

func ExtractRobotsTxt(url string) ([]string, error) {
	hostname, _, err := GetHostname(url)
	if err != nil {
		return []string{}, err
	}
	resp, err := http.Get("https://" + hostname + "/robots.txt")
	defer resp.Body.Close()
	if err != nil {
		return []string{}, err
	}

	disallowedArr := []string{}

	scanner := bufio.NewScanner(resp.Body)
	isAll := false
	for scanner.Scan() {
		line := scanner.Text()
		if line == "User-agent: *" {
			isAll = true
		}
		if isAll && strings.Contains(line, "Disallow:") {
			_, cleanedLine, _ := strings.Cut(line, ": ")
			disallowedArr = append(disallowedArr, cleanedLine)
		}

	}
	if err := scanner.Err(); err != nil {
		return []string{}, err
	}

	return disallowedArr, nil
}
