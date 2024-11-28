const listErrors = document.querySelector("#list-error-popup-container");
const crawlLoader = document.querySelector(".crawl-loader");
const crawlBtn = document.querySelector(".crawl-btn");

function displayErrors(result) {
  listErrors.hidden = false;
  crawlLoader.style.display = "none";
  crawlBtn.style.display = "unset";
  console.log("Error");
  const p = listErrors.children[0];
  const indexedList = listErrors.children[1];
  indexedList.replaceChildren([]);
  p.textContent = result.message;

  // handle network errors
  if (result.data === undefined) {
    p.textContent = "Something went wrong while sending your crawl list.";

    indexedList.replaceChildren(["Please restart the application."]);
    return;
  }

  const sites = result.data.map((site) => {
    const span = document.createElement("span");
    span.textContent = site;
    return span;
  });
  indexedList.replaceChildren(...sites);
}

export default { displayErrors };
