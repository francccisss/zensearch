package utilities

import "testing"

func TestRobotExtraction(t *testing.T) {

	arr, err := ExtractRobotsTxt("https://docs.python.org/")
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(arr) == 0 {
		t.Fatal("Expected a length of greater than 0.")
	}
}
