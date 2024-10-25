package utilities

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
	utilities "web-crawler-service/utilities/links"
)

func ExtractTxt(url string) ([]string, error) {
	hostname, err := utilities.GetHostname(url)
	if err != nil {
		return []string{}, err
	}
	resp, err := http.Get("https://" + hostname + "/robots.txt")
	defer resp.Body.Close()
	if err != nil {
		return []string{}, err
	}

	fmt.Printf("Url: %s\n", resp.Request.URL)
	disallowedArr := make([]string, 10)
	scanner := bufio.NewScanner(resp.Body)
	isAll := false
	for scanner.Scan() {
		line := scanner.Text()
		if line == "User-agent: *" {
			fmt.Println(line)
			isAll = true
		}
		if isAll && strings.Contains(line, "Disallow:") {
			_, cleanedLine, _ := strings.Cut(line, ": ")
			fmt.Println(cleanedLine)
			disallowedArr = append(disallowedArr, cleanedLine)
		}

	}
	if err := scanner.Err(); err != nil {
		return []string{}, err
	}

	return disallowedArr, nil
}
