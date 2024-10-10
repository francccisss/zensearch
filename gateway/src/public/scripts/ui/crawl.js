const listErrors = document.querySelector("#list-error-popup-container");
const crawlLoader = document.querySelector(".crawl-loader");
const crawlBtn = document.querySelector(".crawl-btn");
const crawlSb = document.getElementById("crawl-list-sb");
const template = document.getElementById("waiting-item");

function onCrawlUrls() {
  crawlLoader.style.display = "inline-block";
  crawlBtn.style.display = "none";
  console.log("Request is sent");
  listErrors.hidden = true;
}

function onCrawlDone(result) {
  /*
   Hides the loader ui for crawling after upgrading
   http to wesocket and sending list to the crawler
  */
  console.log("Crawled and Upgraded.");
  crawlLoader.style.display = "none";
  crawlBtn.style.display = "unset";
}

export default { onCrawlDone, onCrawlUrls };
