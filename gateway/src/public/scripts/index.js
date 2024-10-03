import navigation from "./client_navigation/navigation.js";
import crawlInput from "./components/crawl_input/index.js";
import ui from "./ui/index.js";
import pubsub from "./utils/pubsub.js";

const sidebar = document.getElementById("crawl-list-sb");
const openSbBtn = document.getElementById("add-entry-sb-btn");
const crawlBtn = document.querySelector(".crawl-btn");

window.addEventListener("load", () => {
  ui.init();
  navigation.showPage("/");
});

openSbBtn.addEventListener("click", () => {
  sidebar.classList.replace("inactive-sb", "active-sb");
});

sidebar.addEventListener("click", ui.sidebarActions);

function mockPostRequest() {
  const unhiddenInputs = document.querySelectorAll(
    'input.crawl-input:not([data-hidden="true"])',
  );
  const inputValues = Array.from(unhiddenInputs).map((input) => input.value);
  // try catch if an error while sending post request
  try {
    pubsub.publish("crawlStart");
    setTimeout(() => {
      console.log(inputValues);
    }, 3 * 1000);
  } catch (err) {
    // What errors should be thrown?
    // - Entries that already exist on the database
    // - Server error
    // - Value error if its not http format url (TODO handle it client-side too)

    // Result object of type {message:string, status:"error", data: null | unindexedList}
    // result.message will be passed in from err parameter by the try block

    pubsub.publish("crawlError", { status: "error", message: err.message });
    console.error(err.message);
  }
}

crawlBtn.addEventListener("click", mockPostRequest);

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
pubsub.subscribe("crawlStart", () => {
  // poll()
  // in poll function call crawlDone once poll receives
  // response with data.
  const crawlLoader = document.querySelector(".crawl-loader");
  crawlLoader.style.display = "inline-block";
  crawlBtn.style.display = " none";
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
  // This is for handling urls that were already indexed and stored
  // in the database.
});
