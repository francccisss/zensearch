// create a single instance of a crawler
// call crawl method to crawl webpage
// the crawl method will be processed by different threads
// Crawler creates multiple crawl calls for every element in the array

import { Worker } from "worker_threads";

export default class {
  webpages: Array<string> = [];
  constructor(webpages: Array<string>) {
    this.webpages = webpages;
  }
  async start_crawl() {
    // start threadinh here
    this.webpages.forEach((link) => {
      this.crawl(link);
    });
  }

  private async crawl(link: string) {
    console.log(`crawl: ${link}`);
  }
}
