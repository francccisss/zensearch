const listErrors = document.querySelector("#list-error-popup-container");
const crawlLoader = document.querySelector(".crawl-loader");
const crawlBtn = document.querySelector(".crawl-btn");
function onCrawlUrls() {
  crawlLoader.style.display = "inline-block";
  crawlBtn.style.display = "none";
  console.log("Request is sent");
  listErrors.hidden = true;
}

function onCrawlDone(result) {
  console.log("Crawl is done");
  crawlLoader.style.display = "none";
  crawlBtn.style.display = "unset";
}

export default { onCrawlDone, onCrawlUrls };
