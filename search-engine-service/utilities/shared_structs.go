package utilities

type WebpageTFIDF struct {
	Contents    string
	Title       string
	Webpage_url string
	TFScore     float64
	TFIDFRating float64
}

type Webpage struct {
	Contents    string
	Title       string
	Webpage_url string
}
