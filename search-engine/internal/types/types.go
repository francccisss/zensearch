package types

type WebpageTFIDF struct {
	Contents string
	Title    string
	Url      string
	TokenRating
}
type TokenRating struct {
	Bm25rating float64
	TfRating   float64
	IdfRating  float64
}

type WebpageRanking struct {
	Url    string
	Rating float64
}
