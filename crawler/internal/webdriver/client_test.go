package webdriver

import (
	"log"
	"testing"
)

func TestChromeWebDriver(t *testing.T) {

	c, err := CreateClient()
	if err != nil {
		t.Fatalf(err.Error())
	}

	id, err := (*c).NewSession()
	if err != nil {
		t.Fatalf(err.Error())
	}

	log.Println(id)

	err = (*c).Get("https://youtube.com")

	if err != nil {
		t.Fatalf(err.Error())
	}

	(*c).Close()

}
