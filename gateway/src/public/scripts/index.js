import navigation from "./client_navigation/navigation.js";
import crawlInput from "./components/crawl_input/index.js";
import ui from "./ui/index.js";
import cookiesUtil from "./utils/cookies.js";
import pubsub from "./utils/pubsub.js";
import client from "./client_operations/index.js";
import ws from "./client_operations/websocket.js";
import crawl_input from "./components/crawl_input/index.js";

const sidebar = document.getElementById("sidebar-container");
const openSbBtn = document.getElementById("add-entry-sb-btn");
const crawlBtn = document.querySelector(".crawl-btn");
const crawledData = new Map();
let isCrawling = false;

// TODO Add documentations
// TODO store user's unindexed_list in local storage
// for if ever a user refreshes the page while crawling
// we can still redirect users back to the waiting area
// redirect only if there are cookies pertaining a crawling job
// and use the stored list to create the UI for each input
// TODO a crawling acknowledgement, such that when user's receive the crawled website data
// the user can send back an acknoledgement to the websocket server, and only then will the
// websocket server send an `ack` back to the rabbitmq queue.
// TODO Network requests from pubsub needs to be asynchronous
// TODO Handle the errors
// TODO Find a way for the crawler to immediately send message to the
// database service if an error has occured or since it uses a replyqueue to crawl_poll_queue
// let crawler send a message to the express websocket server with an error of the current crawl.
window.addEventListener("load", () => {
  ui.init();
});
openSbBtn.addEventListener("click", () => {
  sidebar.classList.replace("inactive-sb", "active-sb");
});

sidebar.addEventListener("click", async (e) => {
  const target = e.target;
  ui.sidebarActions(e);
  if (target.classList.contains("crawl-btn")) {
    await sendCrawlList();
  }
});

async function sendCrawlList() {
  const unhiddenInputs = document.querySelectorAll(
    'input.crawl-input:not([data-hidden="true"])',
  );
  const inputValues = Array.from(unhiddenInputs).map((input) => input.value);
  try {
    // checkListAndUpgrade returns the list else throws an error and returns null.
    const unindexed_list = await client.checkListAndUpgrade(inputValues);
    console.log("Transition to waiting area for crawled list.");

    const { message_type, job_id } = cookiesUtil.extractCookies();
    // On Successful database check for unindexed list, send the list to the
    // websocket server to start crawling the unindexed list.
    const message = { message_type, unindexed_list, meta: { job_id } };
    ws.send(JSON.stringify(message));
    // might return an error so we need to handle it before we transition
    // to waiting area.

    pubsub.publish("crawlStart", unindexed_list);
  } catch (err) {
    console.error(err.message);
  }
}

/* Pubsub utility is used to handle UI reactivity on data change
 */

// TO SHOW POPUP MESSAGES
pubsub.subscribe("removeEntry", ui.popUpOnRemoveEntry);
pubsub.subscribe("addEntry", ui.popUpOnAddEntry);
// TO SHOW POPUP MESSAGES

pubsub.subscribe("hideEntry", crawlInput.updateEntries);
pubsub.subscribe("revealEntry", crawlInput.updateEntries);
pubsub.subscribe("removeEntry", crawlInput.updateEntries);

pubsub.subscribe("checkAndUpgradeStart", ui.crawlui.onCrawlUrls);
pubsub.subscribe("checkAndUpgradeDone", ui.crawlui.onCrawlDone);
pubsub.subscribe("checkAndUpgradeError", ui.errorsui.handleCrawlErrors);

pubsub.subscribe("crawlReceiver", (msg) => {
  const { job_count } = cookiesUtil.extractCookies();
  const uint8 = new Uint8Array(msg.data_buffer.data);
  const decoder = new TextDecoder();
  const decodedBuffer = decoder.decode(uint8);
  const parseDecodedBuffer = JSON.parse(decodedBuffer);

  console.log(parseDecodedBuffer);
  crawledData.set(parseDecodedBuffer.url, parseDecodedBuffer);
  pubsub.publish("crawlNotify", parseDecodedBuffer);
  if (crawledData.size === Number(job_count)) {
    pubsub.publish("crawlDone", {});
  }
});

// Transition sidebar from crawl list to waiting area
pubsub.subscribe("crawlStart", ui.transitionToWaitingList);

pubsub.subscribe("crawlNotify", (currentCrawledObj) => {
  const waitItems = Array.from(document.querySelectorAll(".wait-item"));
  const updateItems = waitItems.map((waitItem) => {
    const itemText = waitItem.children[0].textContent;
    const url = new URL(itemText);
    if (url.hostname === currentCrawledObj.url) {
      if (currentCrawledObj.message === "Success") {
        waitItem.dataset.state = "done";
        return;
      }
      waitItem.dataset.state = "error";
    }
  });
});

pubsub.subscribe("crawlDone", (currentCrawledObj) => {
  const newListBtn = document.getElementById("new-list-btn");
  newListBtn.style.display = "block";
  // Remove cookies from browser
  cookiesUtil.clearAllCookies();
  console.log("Transition to SEARCH");
});
