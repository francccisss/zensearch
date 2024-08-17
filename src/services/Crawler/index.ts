import { writeFileSync, writeSync } from "fs";
import { userInfo } from "os";
import puppeteer, { Browser, Page, Puppeteer } from "puppeteer";
import { StringDecoder } from "string_decoder";

const remove_hash_url = (link: string) => {
  if (!link.includes("#")) return link;
  const hash_index = link.indexOf("#");
  return link.substring(0, hash_index);
};
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
  webpage_contents: Array<{
    header: { title: string; page_url: string };
    contents: string;
  }>;
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
      webpage_contents: [],
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
    try {
      if (this.page === null) throw new Error("Unable to create browser page.");
      const l = "https://docs.python.org/3/";
      await this.page.goto(l);
      this.data.header.title = await this.page.title();
      await this.traverse_pages(l);
      this.browser?.close();

      writeFileSync("./data.json", JSON.stringify(this.data));
      console.log(this.data);
      console.log(`End of Crawl`);
    } catch (err) {
      const error = err as Error;
      this.browser?.close();
      console.log("LOG: Something went wrong while initializing crawler");
      console.error(error.message);
    } finally {
      console.log("LOG: Close Browser");
      this.browser?.close();
    }
  }
  private async traverse_pages(link: string) {
    const current_page = remove_hash_url(link);
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
      await this.index_page(this.page, current_page);
      const extracted_links = await this.page.$$eval("a", (links) =>
        links.map((link) => link.href),
      );
      const neighbors = remove_duplicates<string>(extracted_links).filter(
        (link) => {
          return link.includes(new URL(current_page).origin) ?? link;
        },
      );
      if (neighbors === undefined || neighbors.length === 0) {
        console.log("LOG: No neighbors.");
        return;
      }
      console.log({
        visited_stack: this.visited_stack,
        current_trace: current_page,
        current_neighbors: neighbors,
      });
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

  private async index_page(current_page: Page, link: string) {
    if (link.includes("#")) {
      const webpage_from_ds = this.data.webpage_contents.find(
        (el: { header: { title: string; page_url: string } }) =>
          remove_hash_url(el.header.page_url) === remove_hash_url(link),
      );
      const is_duplicate = webpage_from_ds !== null;
      if (is_duplicate) {
        console.log("DUPLICATES");
        console.log({
          webpage_from_ds: webpage_from_ds?.header.page_url,
          link,
        });
        return;
      }
    }

    console.log("NOT DUPLICATES");
    console.log({
      link,
    });

    const extract = async () => {
      const filter_data = async (selector: string) => {
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
        return filtered_.join(" ");
      };
      const paras = await filter_data("p");
      const h1 = await filter_data("h1");
      const h2 = await filter_data("h2");
      const h3 = await filter_data("h3");
      const h4 = await filter_data("h4");
      const code = await filter_data("code");
      const pre = await filter_data("pre");
      return await Promise.allSettled([paras, h1, h2, h3, h4, code, pre]);
    };
    const aggregate_data = async (
      extracted_data: PromiseSettledResult<string>[],
    ) => {
      const settle = extracted_data.map((promises) => {
        if (promises.status === "fulfilled") return promises.value;
        return "";
      });

      const l = settle.join("");
      this.data = {
        ...this.data,
        webpage_contents: [
          ...this.data.webpage_contents,
          {
            header: {
              title: (await current_page.title()) || "",
              page_url: (await current_page.url()) || "",
            },
            contents: l,
          },
        ],
      };
    };
    const ex = await extract();
    const ag = await aggregate_data(ex);
  }
}

const scraper = new Scraper();
const crawler = new Crawler(scraper);
crawler.start_crawl(process.argv[2]);
