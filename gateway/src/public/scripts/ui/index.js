import crawlInput from "../components/crawl_input/index.js";

function initCrawlInputs() {
  const listContainer = document.querySelector(".list-container");
  listContainer.appendChild(crawlInput.createComponent());
}

function init() {
  initCrawlInputs();
}

function sidebarActions(event) {
  const sidebar = document.getElementById("crawl-list-sb");
  const target = event.target;
  if (target.classList.contains("new-entry-btn")) {
    crawlInput.addNewEntry();
  }
  if (target.id == "close-sb" && sidebar.classList.contains("active-sb")) {
    sidebar.classList.replace("active-sb", "inactive-sb");
  }
  if (target.classList.contains("remove-entry")) {
    crawlInput.removeEntry(target.dataset.contref);
  }
}

export default { init, sidebarActions };
