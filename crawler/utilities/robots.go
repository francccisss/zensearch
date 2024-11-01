package utilities

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
)

func ExtractRobotsTxt(url string) ([]string, error) {
	hostname, _, err := GetHostname(url)
	if err != nil {
		fmt.Println("ERROR: Unable to get hostname")
		return []string{}, err
	}
	resp, err := http.Get("https://" + hostname + "/robots.txt")
	if err != nil {
		fmt.Println(err.Error())
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
