import navigation from "./client_navigation/navigation.js";
import crawlInput from "./components/crawl_input/index.js";
import ui from "./ui/index.js";
import extract_cookies from "./utils/extract_cookies.js";
import polling from "./utils/polling.js";
import pubsub from "./utils/pubsub.js";
import client from "./client_operations/index.js";

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

crawlBtn.addEventListener("click", async () => {
  const unhiddenInputs = document.querySelectorAll(
    'input.crawl-input:not([data-hidden="true"])',
  );
  const inputValues = Array.from(unhiddenInputs).map((input) => input.value);
  try {
    await client.sendCrawlRequest(inputValues);
    pubsub.publish("crawlDone");
  } catch (err) {
    console.error(err.message);
  }
  // needs to call poll loop after post Request to /crawl
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

pubsub.subscribe("crawlStart", ui.crawlui.onCrawlUrls);
pubsub.subscribe("crawlDone", ui.crawlui.onCrawlDone);
pubsub.subscribe("crawlError", ui.errorsui.handleCrawlErrors);

// ignore this
document.cookie = "job_count=4;";
document.cookie = "job_id=buh;";
extract_cookies();
