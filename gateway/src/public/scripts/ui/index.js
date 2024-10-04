import crawlInput from "../components/crawl_input/index.js";

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
  if (entries.length >= limit) {
    console.log("lol");
    listContainer.appendChild(popup);
    newEntry.disabled = true;
  }
}
function initCrawlInputs() {
  const listContainer = document.querySelector(".list-container");
  listContainer.appendChild(
    crawlInput.createComponent("http://localhost:8080"),
  );
  listContainer.appendChild(
    crawlInput.createComponent("https://fzaid.vercel.app/"),
  );
}

function sidebarActions(event) {
  const sidebar = document.getElementById("crawl-list-sb");
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
}

// TODO create an event listener for crawl button
// need to grab each inputs and for each inputs get their
// values and store them in an array, or unless convert into a form
// then on submit on a form element, we can easily create a form data
// where each input fields can be turned into form entries.

// UI initializer for placeholder or something. idk
function init() {
  initCrawlInputs();
}
export default { init, sidebarActions, popUpOnRemoveEntry, popUpOnAddEntry };
