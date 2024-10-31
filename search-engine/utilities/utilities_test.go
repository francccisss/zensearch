package utilities

import (
	"reflect"
	"testing"
)

func TestTokenizer(t *testing.T) {

	// const query = "Who is raskolnikov in crime and punishment by fyodor dostoevsky?"
	// var expectedOutput = []string{"Who", "is", "raskolnikov", "in", "crime", "and", "punishment", "by", "fyodor", "dostoevsky?"}
	//
	query := "Help"
	var expectedOutput = []string{"Help"}
	tokenizedQuery := Tokenizer(query)
	if !reflect.DeepEqual(expectedOutput, tokenizedQuery) {
		t.Fatalf("Expected output %+v but got %+v", expectedOutput, tokenizedQuery)
	}
}
