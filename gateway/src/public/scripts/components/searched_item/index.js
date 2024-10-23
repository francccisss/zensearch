import uuid from "../../utils/uuid";

const template = document.getElementById("search-item-template");
const searchResultsContainer = document.getElementById(
  "search-results-container",
);

// TODO truncate max length characters for description within the server
// or the search engine.
function createComponent({ link, desc, pageTitle } = searchResult) {
  const newID = uuid();
  const container = document.createElement("div");
  container.append(template.content.cloneNode(true));
  container.setAttribute("class", "searched-item");
  container.querySelector("page-link").href = link;
  const link2 = container.querySelector("page-link2");
  link2.href = link;
  link2.textContent = pageTitle;
  container.querySelector("page-origin").textContent = new URL(link).origin;
  container.querySelector("p").textContent = desc;

  return container;
}

function appendSearchResults(searchResults) {
  const updateResults = searchResults.map((result) => createComponent(result));
  searchResultsContainer.replaceChildren(...updateResults);
}
