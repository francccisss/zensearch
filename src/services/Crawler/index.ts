import puppeteer, { Page, Puppeteer } from "puppeteer";

function remove_duplicates<T>(links: Array<T> | undefined): Array<T> {
  let tmp: Array<T> = [];
  if (links === undefined || links.length === 0) return [];
  outerLoop: for (let i = 0; i < links.length; i++) {
    if (tmp.length === 0) {
      tmp.push(links[i]);
      continue;
    }
    let exists = false;
    for (let j = 0; j < tmp.length; j++) {
      if (links[i] === tmp[j]) {
        exists = true;
        break;
      }
    }
    if (!exists) tmp.push(links[i]);
  }
  return tmp;
}

class Scraper {
  private link: string;
  constructor() {
    this.link = "";
  }
  async launch_browser(): Promise<Page | null> {
    try {
      const browser = await puppeteer.launch({ headless: false });
      const page = await browser.newPage();
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
  private visited_stack: Set<string>;
  private current_browser_page: Page | null;
  constructor(scraper: Scraper) {
    this.scraper = scraper;
    this.visited_stack = new Set<string>([]);
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
    await this.traverse_pages(link);
    console.log(`crawl: ${link}`);
  }
  private async traverse_pages(current_page: string) {
    try {
      if (this.current_browser_page == null)
        throw new Error("Unable to create browser page.");
      await this.current_browser_page.goto(current_page);
      if (this.visited_stack.has(current_page)) {
        console.log("Already Visited: " + current_page);
        return;
      }
      this.visited_stack.add(current_page);

      // FOR CLEANING DATA
      // CHANGE TO SETS HEHE POTAAAAA NO NEED FOR REMOVAL OF DUPS
      const extracted_links = await this.current_browser_page?.$$eval(
        "a",
        (links) => {
          const link_urls = links.map((link) => link.href);
          return link_urls;
        },
      );
      const neighbors = remove_duplicates<string>(extracted_links).filter(
        (link) => {
          if (link.includes("http")) {
            const url = new URL(link);
            return link.includes(current_page) ?? link;
          }
        },
      );
      // FOR CLEANING DATA

      if (neighbors === undefined || neighbors.length === 0) {
        console.log("LOG: End of call.");
        return;
      }

      console.log({
        visited_stack: this.visited_stack,
        current_trace: current_page,
        current_neighbors: neighbors,
      });

      // START RECURSION

      for (let current_neighbor of neighbors) {
        console.log(current_neighbor);
        await this.traverse_pages(current_neighbor);
      }
    } catch (err) {
      const error = err as Error;
      console.log("LOG: Something went wrong when crawling the website");
      console.error(error.message);
    }
  }

  private async index_page() {}
}

const scraper = new Scraper();
const crawler = new Crawler(scraper);
crawler.start_crawl(process.argv[2]);
