import crawlInput from "../components/crawl_input/index.js";

function initCrawlInputs() {
  const listContainer = document.querySelector(".crawl-list-container")
    .children[0];
  listContainer.appendChild(crawlInput.createComponent());
}

function init() {
  initCrawlInputs();
}

export default { init };
