package utils

func stringToSlice(str string) []string {
	tmp := []string{}
	startP := 0
	for i := range str {
		if str[i] == ' ' {
			tmp = append(tmp, str[startP:i])
			startP = i + 1
		}
	}
	tmp = append(tmp, str[startP:])
	return tmp
}
