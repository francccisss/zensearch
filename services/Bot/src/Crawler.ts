import puppeteer, { Browser, Page } from "puppeteer";
import remove_duplicate from "./utils/remove_duplicate";

const remove_hash_url = (link: string) => {
  if (!link.includes("#")) return link;
  const hash_index = link.indexOf("#");
  return link.substring(0, hash_index);
};

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
    }
  }

  set(new_link: string) {
    this.link = new_link;
    return new_link;
  }
}
type data_t = {
  webpages: Array<{
    header: { title: string; webpage_url: string };
    contents: string;
  }>;
  header: {
    title: string;
    url: string;
  };
};

class Crawler {
  private scraper: Scraper;
  private visited_stack: Set<string>;
  private browser: Browser | null;
  private page: Page | null;
  private max_interval: number = 5000;
  private min_interval: number = 1500;
  private current_interval: number = 500;
  data: data_t;

  constructor(scraper: Scraper) {
    this.scraper = scraper;
    this.visited_stack = new Set<string>([]);
    this.browser = null;
    this.page = null;
    this.data = {
      header: {
        title: "",
        url: "",
      },
      webpages: [],
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
      await this.page.goto(link);
      this.data.header = { title: await this.page.title(), url: link };
      await this.traverse_pages(link);
      this.browser?.close();
      console.log(`End of Crawl`);
      return this.data;
    } catch (err) {
      const error = err as Error;
      this.browser?.close();
      console.log("LOG: Something went wrong while initializing crawler");
      console.error(error.message);
      return this.data;
    } finally {
      console.log("LOG: Close Browser");
      this.browser?.close();
    }
  }

  private async throttle_visits() {
    const DEC_INC = 70;
    if (
      this.current_interval < this.min_interval ||
      this.current_interval < this.max_interval
    ) {
      this.current_interval += DEC_INC;
    } else if (this.current_interval > this.max_interval) {
      this.current_interval = 500;
    }
    console.log({ current_interval: this.current_interval });
    return new Promise((resolve) => {
      setTimeout(async () => {
        resolve("GO NEXT");
      }, this.current_interval);
    });
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
      await this.throttle_visits();
      await this.page.goto(current_page);
      await this.index_page(this.page, current_page);
      const css_selector =
        'a:not([href$=".zip"]):not([href$=".pdf"]):not([href$=".exe"]):not([href$=".jpg"]):not([href$=".png"]):not([href$=".tar.gz"]):not([href$=".rar"]):not([href$=".7z"]):not([href$=".mp3"]):not([href$=".mp4"]):not([href$=".mkv"]):not([href$=".tar"]):not([href$=".xz"]):not([href$=".msi"])';
      const extracted_links = await this.page.$$eval(css_selector, (links) =>
        links.map((link) => link.href),
      );
      const neighbors = remove_duplicate<string>(extracted_links).filter(
        (link) => {
          return link.includes(new URL(current_page).origin) ?? link;
        },
      );
      if (neighbors === undefined || neighbors.length === 0) {
        console.log("LOG: No neighbors.");
        return;
      }
      console.log({
        visited_stack: this.visited_stack.size,
        current_trace: current_page,
        current_neighbors: neighbors.length,
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
      const webpage_from_ds = this.data.webpages.find(
        (el: { header: { title: string; webpage_url: string } }) =>
          remove_hash_url(el.header.webpage_url) === remove_hash_url(link),
      );
      const is_duplicate = webpage_from_ds !== null;
      if (is_duplicate) {
        console.log("DUPLICATES");
        console.log({
          webpage_from_ds: webpage_from_ds!.header.webpage_url,
          link,
        });
        return;
      }
    }

    console.log("NOT DUPLICATES");
    console.log({ link });

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
        webpages: [
          ...this.data.webpages,
          {
            header: {
              title: (await current_page.title()) || "",
              webpage_url: current_page.url() || "",
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

export { Scraper, Crawler };
