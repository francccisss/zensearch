package utilities

import "testing"

func TestRobotExtraction(t *testing.T) {

	arr, err := ExtractTxt("https://fzaid.vercel.app/")
	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(arr) == 0 {
		t.Fatalf("Expected a length of greater than 0.")
	}
}
