import pubsub from "../../utils/pubsub.js";
import uuid from "../../utils/uuid.js";

const template = document.getElementById("crawl-input-template");
const listContainer = document.querySelector(".list-container");

function createComponent() {
  const newId = uuid();
  const container = document.createElement("div");
  container.append(template.content.cloneNode(true));
  const btns = container.querySelectorAll("button");
  btns.forEach((btn) => btn.setAttribute("data-contref", newId));
  container.classList.add("url-entry");
  container.classList.add("reveal-entry");
  container.setAttribute("id", newId);
  console.log(container);
  return container;
}

function addNewEntry() {
  listContainer.appendChild(createComponent());
  pubsub.publish("addEntry", listContainer.children);
}

function hideEntry(ref) {
  const entries = Array.from(listContainer.children);
  const updatedEntries = entries.map((child) => {
    if (child.id === ref) {
      child.classList.replace("reveal-entry", "hide-entry");
      child.dataset.hidden = true;
      child.querySelector("input").disabled = true;
      return child;
    }
    return child;
  });
  pubsub.publish("hideEntry", updatedEntries);
}
function revealEntry(ref) {
  const entries = Array.from(listContainer.children);
  const updatedEntries = entries.map((child) => {
    if (child.id === ref) {
      child.classList.replace("hide-entry", "reveal-entry");
      child.dataset.hidden = false;
      child.querySelector("input").disabled = false;
      return child;
    }
    return child;
  });
  pubsub.publish("revealEntry", updatedEntries);
}

/*
  Params `ref` reference to the input we want to remove from the
  crawl list.
 */
function removeEntry(ref) {
  const entries = Array.from(listContainer.children);
  const filtered = entries.filter((child) => child.id !== ref ?? child);
  pubsub.publish("removeEntry", filtered);
}

/*
  Params`newEntries` refers to the updated entries for deleting through pubsub
*/
function updateEntries(newEntries) {
  if (newEntries !== null) {
    listContainer.replaceChildren(...newEntries);
  }
  console.log(listContainer);
}

export default {
  createComponent,
  addNewEntry,
  removeEntry,
  updateEntries,
  revealEntry,
  hideEntry,
};
