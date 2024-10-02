import crawlInput from "../components/crawl_input/index.js";

function initCrawlInputs() {
  const listContainer = document.querySelector(".crawl-list-container")
    .children[0];
  listContainer.appendChild(crawlInput.createComponent());
}

function init() {
  initCrawlInputs();
}

function sidebarActions(event) {
  const sidebar = document.getElementById("crawl-list-sb");
  if (event.target.classList.contains("new-entry-btn")) {
    crawlInput.addNewEntry();
  }
  if (
    event.target.id == "close-sb" &&
    sidebar.classList.contains("active-sb")
  ) {
    sidebar.classList.replace("active-sb", "inactive-sb");
  }
  if (event.target.id == "remove-entry") {
  }
}

export default { init, sidebarActions };
