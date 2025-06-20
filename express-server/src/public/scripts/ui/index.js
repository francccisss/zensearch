import crawlInput from "../components/crawl_input/index.js";
import errorsui from "./errors.js";
import crawlui from "./crawl.js";
import cookiesUtil from "../utils/cookies.js";

/*
 * The plan is to make this file as a collection of ui component/element handlers
 * instead of spreading them out wherever, or putting them in the same folder
 * with the entry point `index.js` where all of the event listeners reside.
 */

const newEntry = document.querySelector(".new-entry-btn");
const listContainer = document.querySelector(".list-container");
const popup = document.createElement("p");
const limit = 6;

popup.classList.add("info-large");

function popUpOnRemoveEntry(entries) {
  if (entries.length < limit) {
    newEntry.disabled = false;
    const children = Array.from(listContainer.children);
    children.forEach((child) => {
      if (child.classList.contains("info-large")) {
        child.remove();
      }
    });
  }
}

function popUpOnAddEntry(entries) {
  popup.textContent = "You've reached the maximum limit.";
  if (entries.length == limit) {
    listContainer.appendChild(popup);
    newEntry.disabled = true;
  }
}
function initCrawlInputs() {
  const listContainer = document.querySelector("#crawl-list-container");
  listContainer.appendChild(crawlInput.createComponent(""));
}

function sidebarActions(event) {
  const sidebar = document.getElementById("sidebar-container");
  const target = event.target;
  if (target.classList.contains("new-entry-btn")) {
    crawlInput.addNewEntry();
  }
  if (target.id == "close-sb-btn" && sidebar.classList.contains("active-sb")) {
    sidebar.classList.replace("active-sb", "inactive-sb");
  }
  if (target.classList.contains("remove-entry-btn")) {
    crawlInput.removeEntry(target.dataset.contref);
  }

  if (target.classList.contains("hide-reveal-entry-btn")) {
    if (target.classList.contains("reveal-entry-btn")) {
      target.classList.replace("reveal-entry-btn", "hide-entry-btn");
      crawlInput.hideEntry(target.dataset.contref);
      return;
    }
    target.classList.replace("hide-entry-btn", "reveal-entry-btn");
    crawlInput.revealEntry(target.dataset.contref);
  }
  if (target.id === "new-list-btn") {
    transitionToCrawlList([]);
  }
}

function transitionToCrawlList() {
  const crawlList = document.getElementById("crawl-list-sb");
  const waitingList = document.getElementById("waiting-list-sb");
  const newListBtn = document.getElementById("new-list-btn");
  if (crawlList.classList.contains("inactive-list-container")) {
    waitingList.classList.replace(
      "active-list-container",
      "inactive-list-container",
    );
    crawlList.classList.replace(
      "inactive-list-container",
      "active-list-container",
    );
  }
  newListBtn.style.display = "none";
}
function transitionToWaitingList(list) {
  const crawlList = document.getElementById("crawl-list-sb");
  const waitingList = document.getElementById("waiting-list-sb");
  if (crawlList.classList.contains("active-list-container")) {
    crawlList.classList.replace(
      "active-list-container",
      "inactive-list-container",
    );
    waitingList.classList.replace(
      "inactive-list-container",
      "active-list-container",
    );
  }
  const template = document.getElementById("waiting-item");
  const waitListContainer = document.getElementById("wait-list-container");
  const createItems = list.map((item) => {
    const container = document.createElement("div");
    container.append(template.content.cloneNode(true));
    container.setAttribute("class", "wait-item");
    container.dataset.state = item.state;
    const p = container.children[0];
    const icon = container.children[1];
    p.textContent = item.url;
    return container;
  });
  waitListContainer.replaceChildren(...createItems);
}
function init() {
  if (document.cookie == "") {
    initCrawlInputs();
    return;
  }
  const { message_type } = cookiesUtil.extractCookies();
  if (message_type == "crawling") {
    const list = JSON.parse(localStorage.getItem("list"));
    if (
      list.every((item) => item.state === "done") ||
      list.every((item) => item.state === "error")
    ) {
      cookiesUtil.clearAllCookies();
      localStorage.clear();
      return;
    }
    transitionToWaitingList(JSON.parse(localStorage.getItem("list")));
    return;
  }
}

export default {
  errorsui,
  crawlui,
  init,
  sidebarActions,
  popUpOnRemoveEntry,
  popUpOnAddEntry,
  transitionToWaitingList,
};
