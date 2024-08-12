class Crawler {
  webpage: string;
  constructor(webpage: string) {
    this.webpage = webpage;
  }

  async start_crawl() {
    this.crawl(this.webpage);
  }

  private async crawl(link: string) {
    console.log(`crawl: ${link}`);
  }
}

const crawler = new Crawler(process.argv[2]);
crawler.start_crawl();
