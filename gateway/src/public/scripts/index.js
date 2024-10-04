import navigation from "./client_navigation/navigation.js";
import crawlInput from "./components/crawl_input/index.js";
import ui from "./ui/index.js";
import extract_cookies from "./utils/extract_cookies.js";
import polling from "./utils/polling.js";
import pubsub from "./utils/pubsub.js";

const crawlLoader = document.querySelector(".crawl-loader");
const sidebar = document.getElementById("crawl-list-sb");
const openSbBtn = document.getElementById("add-entry-sb-btn");
const crawlBtn = document.querySelector(".crawl-btn");
const listErrors = document.querySelector("#list-error-popup-container");

window.addEventListener("load", () => {
  ui.init();
  navigation.showPage("/");
});

openSbBtn.addEventListener("click", () => {
  sidebar.classList.replace("inactive-sb", "active-sb");
});

sidebar.addEventListener("click", ui.sidebarActions);

async function mockPostRequest(webUrls) {
  // try catch if an error while sending post request
  try {
    pubsub.publish("crawlStart");
    const sendWebUrls = await fetch("http://localhost:8080/crawl", {
      mode: "cors",
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(webUrls),
    });
    const response = await sendWebUrls.json();
    console.log(response);
    if (response.is_crawling === false) {
      throw new Error(JSON.stringify(response));
    }
    // needs to call poll loop
  } catch (err) {
    const parseErr = JSON.parse(err.message);
    // What errors should be thrown?
    // - Entries that already exist on the database
    // - Server error
    // - Value error if its not http format url (TODO handle it client-side too)

    // Result object of type {message:string, status:"error", data: null | unindexedList}
    // result.message will be passed in from err parameter by the try block

    pubsub.publish("crawlError", {
      status: "error",
      message: parseErr.message,
      data: parseErr.crawl_list,
    });
    console.error(parseErr.message);
  }
}

crawlBtn.addEventListener("click", async () => {
  const unhiddenInputs = document.querySelectorAll(
    'input.crawl-input:not([data-hidden="true"])',
  );
  const inputValues = Array.from(unhiddenInputs).map((input) => input.value);
  await mockPostRequest(inputValues);
});

/* Pubsub utility is used to handle UI reactivity on data change
 */

// TO SHOW POPUP MESSAGES
pubsub.subscribe("removeEntry", ui.popUpOnRemoveEntry);
pubsub.subscribe("addEntry", ui.popUpOnAddEntry);
// TO SHOW POPUP MESSAGES

pubsub.subscribe("hideEntry", crawlInput.updateEntries);
pubsub.subscribe("revealEntry", crawlInput.updateEntries);
pubsub.subscribe("removeEntry", crawlInput.updateEntries);

// on send post request to server with provided webUrls
// what happens if a post request is sent to the server.

// polling is started here.
pubsub.subscribe("crawlStart", async () => {
  crawlLoader.style.display = "inline-block";
  crawlBtn.style.display = "none";
  console.log("Request is sent");
});

pubsub.subscribe("crawlDone", (result) => {
  console.log("Crawl is done");
  crawlLoader.style.display = "none";
  crawlBtn.style.display = "unset";
});

pubsub.subscribe("crawlError", (result) => {
  // This is a standard error for server response
  // - Wrong values
  // - Server error
  crawlLoader.style.display = "none";
  crawlBtn.style.display = "unset";
  if (result.data == null) {
    console.log(result.message);
    return;
  }
  console.log("Error");
  const p = listErrors.children[0];
  const indexedList = listErrors.children[1];
  p.textContent = result.message;

  const sites = result.data.map((site) => {
    const span = document.createElement("span");
    span.textContent = site;
    return span;
  });
  indexedList.replaceChildren(...sites);
  // This is for handling urls that were already indexed and stored
  // in the database.
});

document.cookie = "job_count=4;";
document.cookie = "job_id=buh;";
extract_cookies();
