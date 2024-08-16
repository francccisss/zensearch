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

type data_t = {
  [key: string]: any;
  header: {
    title: string;
  };
};
class Crawler {
  private scraper: Scraper;
  private visited_stack: Set<string>;
  private browser: Browser | null;
  private page: Page | null;
  private stack_frame_count: number;
  data: data_t;

  constructor(scraper: Scraper) {
    this.scraper = scraper;
    this.visited_stack = new Set<string>([]);
    this.browser = null;
    this.page = null;
    this.stack_frame_count = 0;
    this.data = {
      header: {
        title: "",
      },
      p: [],
      h1: [],
      h2: [],
      h3: [],
      h4: [],
      code: [],
      pre: [],
    };
  }
  async start_crawl(link: string) {
    try {
      this.scraper.set(link);
      this.browser = await this.scraper.launch_browser();
      if (this.browser == null)
        throw new Error("Unable to create browser page.");

      this.page = await this.browser.newPage();
      if (this.page === null) throw new Error("Unable to create browser page.");
      await this.crawl(link);
    } catch (err) {
      const error = err as Error;
      this.browser?.close();
      console.log("LOG: Something went wrong with the crawler.");
      console.error(error.message);
    }
  }

  private async crawl(link: string) {
    if (this.page === null) throw new Error("Unable to create browser page.");
    await this.page.goto(link);
    this.data.header.title = await this.page.title();
    await this.traverse_pages(link);
    this.browser?.close();
    console.log(this.data);
    console.log(`End of Crawl`);
  }
  private async traverse_pages(current_page: string) {
    try {
      if (this.browser == null)
        throw new Error("Unable to create browser page.");
      if (this.page == null) throw new Error("Unable to create browser page.");

      if (this.visited_stack.has(current_page)) {
        console.log("Already Visited: " + current_page);
        return this.page;
      }
      this.visited_stack.add(current_page);

      await this.page.goto(current_page);

      // INDEX CURRENT PAGE
      await this.index_page(this.page);

      // FOR CLEANING DATA
      // CHANGE TO SETS HEHE POTAAAAA NO NEED FOR REMOVAL OF DUPS
      const extracted_links = await this.page.$$eval("a", (links) =>
        links.map((link) => link.href),
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

  private async index_page(current_page: Page) {
    const extract = async (selector: string) => {
      const map_ = await current_page.$$eval(selector, (el) =>
        el.map((p) => {
          return p.textContent;
        }),
      );
      const filtered_ = map_.filter((p) => {
        if (p !== undefined) {
          return p;
        }
      });
      this.data = {
        ...this.data,
        [selector]: [...(this.data[selector] as any[]), ...filtered_],
      };
    };

    const paras = await extract("p");
    const h1 = await extract("h1");
    const h2 = await extract("h2");
    const h3 = await extract("h3");
    const h4 = await extract("h4");
    const code = await extract("code");
    const pre = await extract("pre");
    //this.data = {
    //  header: { ...this.data.header },
    //  paras: [...this.data.paras, ...paras],
    //  h1: [...this.data.h1, ...h1],
    //  h2: [...this.data.h2, ...h2],
    //  h3: [...this.data.h3, ...h3],
    //  h4: [...this.data.h4, ...h4],
    //  code: [...this.data.code, ...code],
    //  pre: [...this.data.pre, ...pre],
    //};
  }
}

const scraper = new Scraper();
const crawler = new Crawler(scraper);
crawler.start_crawl(process.argv[2]);
