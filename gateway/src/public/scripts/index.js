import crawlInput from "./components/crawl_input/index.js";
const pageUrls = {
  home: "/",
  crawlProcess: "/crawl/process",
  results: "/search/results",
};

function showPage(path) {
  const pagesEl = document.querySelectorAll('main[id*="-page"]');
  pagesEl.forEach((page) => {
    page.hidden = true;
  });
  if (path === pageUrls.home) {
    document.getElementById("search-page").hidden = false;
  } else if (path === pageUrls.results) {
    document.getElementById("results-page").hidden = false;
  } else if (path === pageUrls.crawlProcess) {
    document.getElementById("process-page").hidden = false;
  }
}

window.addEventListener("load", () => {
  showPage("/");
});
