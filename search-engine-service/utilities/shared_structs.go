package utilities

type WebpageTFIDF struct {
	Contents   string
	Title      string
	Url        string
	TFScore    float64
	BM25Rating float64
}

type Webpage struct {
	Contents string
	Title    string
	Url      string
}

var webpages = []Webpage{
	{
		Contents: "Welcome to the homepage!",
		Title:    "Home",
		Url:      "https://example.com/home",
	},
	{
		Contents: "Learn more about our services and offerings.",
		Title:    "About Us",
		Url:      "https://example.com/about",
	},
	{
		Contents: "Contact us via email or phone.",
		Title:    "Contact Us",
		Url:      "https://example.com/contact",
	},
	{
		Contents: "Our blog covers the latest tech trends.",
		Title:    "Tech Blog",
		Url:      "https://example.com/blog/tech",
	},
	{
		Contents: "Checkout our products and offers.",
		Title:    "Products",
		Url:      "https://example.com/products",
	},
	{
		Contents: "Find solutions to your common issues.",
		Title:    "Support",
		Url:      "https://example.com/support",
	},
	{
		Contents: "Download our app for better experience.",
		Title:    "Download",
		Url:      "https://example.com/download",
	},
	{
		Contents: "Latest news in the world of tech.",
		Title:    "Tech News",
		Url:      "https://example.com/news/tech",
	},
	{
		Contents: "Join our community forum.",
		Title:    "Forum",
		Url:      "https://example.com/forum",
	},
	{
		Contents: "Read our privacy policy and terms.",
		Title:    "Privacy Policy",
		Url:      "https://example.com/privacy",
	},
}
