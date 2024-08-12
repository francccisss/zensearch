import puppeteer, { Page, Puppeteer } from "puppeteer";

class Scraper {
  private link: string;
  constructor() {
    this.link = "";
  }
  async launch_browser(): Promise<Page | null> {
    try {
      const browser = await puppeteer.launch({ headless: false });
      const page = await browser.newPage();
      await page.goto(this.link);
      const test = await page.$(".banner");
      console.log(test);
      return page;
    } catch (err) {
      const error = err as Error;
      console.log("LOG: Browser closed unexpectedly.");
      console.error(error.message);
      process.exit(1);
      return null;
    }
  }

  set(new_link: string) {
    this.link = new_link;
    return new_link;
  }
}

// when index_page is triggered, store processed data on a data structure

class Crawler {
  private scraper: Scraper;
  private visited_stack: Array<string>;
  private current_browser_page: Page | null;
  constructor(scraper: Scraper) {
    this.scraper = scraper;
    this.visited_stack = [];
    this.current_browser_page = null;
  }
  async start_crawl(link: string) {
    try {
      this.scraper.set(link);
      this.current_browser_page = await this.scraper.launch_browser();
      if (this.current_browser_page == null)
        throw new Error("Unable to create browser page.");
      this.crawl(link);
    } catch (err) {
      const error = err as Error;
      console.log("LOG: Something went wrong with the crawler.");
      console.error(error.message);
    }
  }

  private async crawl(link: string) {
    // grab all of the links from current page (link)
    const neighbors = ["link1", "link2", "link3"];
    this.traverse_pages(link, neighbors); // need to keep feeding new neighbors
    console.log(`crawl: ${link}`);
  }

  private traverse_pages(current_page: string, link_array: Array<string>) {
    // head of the stack
    // comparing strings
    if (current_page !== this.visited_stack[0]) {
      this.visited_stack.push(current_page);
    }
  }

  private async index_page() {}
}

const scraper = new Scraper();
const crawler = new Crawler(scraper);
crawler.start_crawl(process.argv[2]);
