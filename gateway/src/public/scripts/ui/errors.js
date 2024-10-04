const listErrors = document.querySelector("#list-error-popup-container");
const crawlLoader = document.querySelector(".crawl-loader");
const crawlBtn = document.querySelector(".crawl-btn");

function handleCrawlErrors(result) {
  // This is a standard error for server response
  // - Wrong values
  // - Server error
  listErrors.hidden = false;
  crawlLoader.style.display = "none";
  crawlBtn.style.display = "unset";
  console.log("Error");
  const p = listErrors.children[0];
  const indexedList = listErrors.children[1];
  indexedList.replaceChildren([]);
  p.textContent = result.message;

  // handle network errors
  if (result.data == undefined) {
    console.log(result.message);
    p.textContent = "Something went wrong while sending your crawl list.";

    indexedList.replaceChildren(["Please restart the application."]);
    return;
  }

  // This is for handling urls that were already indexed and stored
  // in the database.
  const sites = result.data.map((site) => {
    const span = document.createElement("span");
    span.textContent = site;
    return span;
  });
  indexedList.replaceChildren(...sites);
}

export default { handleCrawlErrors };
