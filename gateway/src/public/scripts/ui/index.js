import crawlInput from "../components/crawl_input/index.js";

/*
 * The plan is to make this file as a collection of ui component/element handlers
 * instead of spreading them out wherever, or putting them in the same folder
 * with the entry point `index.js` where all of the event listeners reside.
 */

function initCrawlInputs() {
  const listContainer = document.querySelector(".list-container");
  listContainer.appendChild(crawlInput.createComponent());
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

// UI initializer for placeholder or something. idk
function init() {
  initCrawlInputs();
}
export default { init, sidebarActions };
