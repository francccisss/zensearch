import pubsub from "../../utils/pubsub.js";
import uuid from "../../utils/uuid.js";

const template = document.getElementById("crawl-input-template");
const crawlListContainer = document.querySelector("#crawl-list-container");

/* File for handling the input entries for creating new crawl entries
 * of url for crawler to work with.
 *
 */
function createComponent(url) {
  const newId = uuid();
  const container = document.createElement("div");
  container.append(template.content.cloneNode(true));
  const btns = container.querySelectorAll("button");
  btns.forEach((btn) => btn.setAttribute("data-contref", newId));
  container.classList.add("url-entry");
  container.classList.add("reveal-entry");
  container.setAttribute("id", newId);
  console.log(container);

  const input = container.querySelector("input");
  input.value = url;
  return container;
}

function addNewEntry() {
  crawlListContainer.appendChild(createComponent(""));
  pubsub.publish("addEntry", crawlListContainer.children);
}

function hideEntry(ref) {
  const entries = Array.from(document.querySelectorAll(".url-entry"));
  const updatedEntries = entries.map((entry) => {
    if (entry.id === ref) {
      entry.classList.replace("reveal-entry", "hide-entry");
      entry.dataset.hidden = true;

      const input = entry.querySelector("input");
      input.disabled = true;
      input.setAttribute("data-hidden", "true");
      return entry;
    }
    return entry;
  });
  pubsub.publish("hideEntry", updatedEntries);
}
function revealEntry(ref) {
  const entries = Array.from(document.querySelectorAll(".url-entry"));
  const updatedEntries = entries.map((entry) => {
    if (entry.id === ref) {
      entry.classList.replace("hide-entry", "reveal-entry");
      entry.dataset.hidden = false;

      const input = entry.querySelector("input");
      input.disabled = false;
      input.setAttribute("data-hidden", "false");
      return entry;
    }
    return entry;
  });
  pubsub.publish("revealEntry", updatedEntries);
}

/*
  Params `ref` reference to the input we want to remove from the
  crawl list.
 */
function removeEntry(ref) {
  const entries = Array.from(document.querySelectorAll(".url-entry"));
  console.log(entries.length);
  if (entries.length < 2) {
    return;
  }
  const filtered = entries.filter((child) => child.id !== ref ?? child);
  pubsub.publish("removeEntry", filtered);
}

/*
  Params`newEntries` refers to the updated entries for deleting through pubsub
*/
function updateEntries(newEntries) {
  if (newEntries !== null) {
    crawlListContainer.replaceChildren(...newEntries);
  }
  console.log(crawlListContainer);
}

function submitCrawlList() {}

export default {
  createComponent,
  addNewEntry,
  removeEntry,
  updateEntries,
  revealEntry,
  hideEntry,
  submitCrawlList,
};
