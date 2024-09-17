package crawler

import (
	"context"
	"fmt"
)

func Crawl(ctx context.Context, w string) {
	fmt.Printf("Start Crawling %s\n", w)
}
