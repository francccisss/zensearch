import crawlInput from "./components/crawl_input/index.js";
import ui from "./ui/index.js";
import pubsub from "./utils/pubsub.js";

const sidebar = document.getElementById("crawl-list-sb");
const closeSbBtn = document.getElementById("close-sb");
const openSbBtn = document.getElementById("add-entry-sb-btn");
const listContainer = document.querySelector(".list-container");
const pageUrls = {
  home: "/",
  crawlProcess: "#crawl-section/process",
  results: "/search/results",
};

function showPage(path) {
  const pagesEl = document.querySelectorAll('main[id*="-page"]');
  pagesEl.forEach((page) => {
    page.hidden = true;
    page.style.display = "none";
  });
  if (path === pageUrls.home) {
    console.log(pageUrls.home);
    document.getElementById("search-page").hidden = false;
    document.getElementById("search-page").style.display = "flex";
  } else if (path === pageUrls.results) {
    console.log(pageurls.home);
    document.getElementById("results-page").hidden = false;
  } else {
    console.log(pageUrls.home);
    document.getElementById("process-page").hidden = false;
  }
}

window.addEventListener("load", () => {
  ui.init();
  showPage("/");
});

openSbBtn.addEventListener("click", () => {
  sidebar.classList.replace("inactive-sb", "active-sb");
});

sidebar.addEventListener("click", ui.sidebarActions);
pubsub.subscribe("removeEntry", crawlInput.updateEntries);
