import { userInfo } from "os";
import puppeteer, { Browser, Page, Puppeteer } from "puppeteer";
import { StringDecoder } from "string_decoder";

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
  async launch_browser(): Promise<Browser | null> {
    try {
      const browser = await puppeteer.launch();
      return browser;
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
  private browser: Browser | null;
  private page: Page | null;
  constructor(scraper: Scraper) {
    this.scraper = scraper;
    this.visited_stack = new Set<string>([]);
    this.browser = null;
    this.page = null;
  }
  async start_crawl(link: string) {
    try {
      this.scraper.set(link);
      this.browser = await this.scraper.launch_browser();
      if (this.browser == null)
        throw new Error("Unable to create browser page.");

      this.page = await this.browser.newPage();
      await this.crawl(link);
    } catch (err) {
      const error = err as Error;
      this.browser?.close();
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
      if (this.browser == null)
        throw new Error("Unable to create browser page.");
      if (this.page == null) throw new Error("Unable to create browser page.");

      if (this.visited_stack.has(current_page)) {
        console.log("Already Visited: " + current_page);
        return;
      }
      this.visited_stack.add(current_page);

      await this.page.goto(current_page);
      // FOR CLEANING DATA
      // CHANGE TO SETS HEHE POTAAAAA NO NEED FOR REMOVAL OF DUPS
      const extracted_links = await this.page.$$eval("a", (links) => {
        const link_urls = links.map((link) => link.href);
        return link_urls;
      });
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
        console.log("LOG: No neighbors.");
        return;
      }

      console.log({
        visited_stack: this.visited_stack,
        current_trace: current_page,
        current_neighbors: neighbors,
      });

      // START RECURSION

      for (let current_neighbor of neighbors) {
        await this.traverse_pages(current_neighbor);
      }
    } catch (err) {
      const error = err as Error;
      this.browser?.close();
      console.log("LOG: Something went wrong when crawling the website");
      console.error(error.message);
    }
  }

  private async index_page() {
    //How does indexing work?
    //To analyse a page and determine if it contains meaningful content for indexing.
    //meaningful content depends on how we want it to be:
    //  How do we determine if it is meaningful in this case?
    //    User queries are used to look up every indexed page in the database
    //    for retrieval to be served back to the user, a user query is just a string,
    //    this string is used to look up the database for relevant pages from the query string
    //
    //    in a search engine, meaningful content means every content that **might** be useful
    //    to the users. what are meaningful to the users? the contents on the page and what
    //    are the contents on the page? strings, images, videos, but in this case, we'll
    //    simplify the process and use the strings in a page as the meaningful content,
    //    but how much data that needs to be extract from the webpage? and how do we determine within all
    //    of the strings on a webpage that is meaningful, because we can't just take every string, some strings on
    //    every webpage might contain duplicates, like a footer that is appended on every page, should that be meaningful?
    //    so we need to list down what are considered to be meaningful contents for this search engine.
    //
    //
    //  - Create a list of elements that we need to extract from a page
    //    - On first visit, extract the header at the beginning
    //      ## List of elements to extract from body element
    //        - Headers from h1 to h6
    //        - P
    //        - Image alt
    //        - Code
    //        - Pre
    //
    //  After distinguishing meaningful content within a webpage for indexing
    //  now we need a way extract these content and store in a data structure,
    //  but first lets figure out what data we need to use to store the indexed webpage:
    //    - Needs to be atleast O(log N) | O(1) | O(N) for retrieving.
    //    - Needs to be O(N) to look up each element for processing in the serving stage.
    //    - Deletion wont matter... yet.
  }
}

const scraper = new Scraper();
const crawler = new Crawler(scraper);
crawler.start_crawl(process.argv[2]);
