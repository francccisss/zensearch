import puppeteer, { Puppeteer } from "puppeteer";

class Scraper {
  private link: string;
  constructor() {
    this.link = "";
  }
  async launch_browser() {
    const browser = await puppeteer.launch({ headless: false });
    const page = await browser.newPage();
    await page.goto(this.link);
    const test = await page.$(".banner");
    console.log(test);
  }
  set(new_link: string) {
    this.link = new_link;
    return new_link;
  }
}

// when index_page is triggered, store processed data on a data structure

class Crawler {
  private scraper: Scraper;
  constructor(Scraper: { new (): Scraper }) {
    this.scraper = new Scraper();
  }
  async start_crawl(link: string) {
    this.crawl(link);
  }
  private async crawl(link: string) {
    this.scraper.set(link);
    this.scraper.launch_browser();
    console.log(`crawl: ${link}`);
  }
  private async index_page() {}
}

const crawler = new Crawler(Scraper);
crawler.start_crawl(process.argv[2]);
