import crawlInput from "./components/crawl_input/index.js";
import ui from "./ui/index.js";
import cookiesUtil from "./utils/cookies.js";
import pubsub from "./utils/pubsub.js";
import client from "./client/index.js";
import clientws from "./client/websocket.js";

const sidebar = document.getElementById("sidebar-container");
const openSbBtn = document.getElementById("add-entry-sb-btn");
const crawledData = new Map();

// TODO Add documentations
// TODO Network requests from pubsub needs to be asynchronous
// TODO Handle the errors
window.addEventListener("load", ui.init);
openSbBtn.addEventListener("click", () => {
  sidebar.classList.replace("inactive-sb", "active-sb");
});

sidebar.addEventListener("click", async (e) => {
  const target = e.target;
  ui.sidebarActions(e);
  if (target.classList.contains("crawl-btn")) {
    await client.sendCrawlList();
  }
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

pubsub.subscribe("checkAndUpgradeStart", ui.crawlui.onCrawlUrls);
pubsub.subscribe("checkAndUpgradeDone", ui.crawlui.onCrawlDone);
pubsub.subscribe("checkAndUpgradeError", ui.errorsui.displayErrors);

// Parses incoming message from the message queue via websocket message
pubsub.subscribe("crawlReceiver", (msg) => {
  const { job_count } = cookiesUtil.extractCookies();
  const uint8 = new Uint8Array(msg.data_buffer.data);
  const decoder = new TextDecoder();
  const decodedBuffer = decoder.decode(uint8);
  const parseDecodedBuffer = JSON.parse(decodedBuffer);

  crawledData.set(parseDecodedBuffer.URLSeed, parseDecodedBuffer);
  // sends an ack back to the server that the crawl message status was received by client
  clientws.ackMessage();
  pubsub.publish("crawlNotify", parseDecodedBuffer);
  if (crawledData.size === Number(job_count)) {
    pubsub.publish("crawlDone", {});
  }
});

// Transition sidebar from crawl list to waiting area
pubsub.subscribe("crawlStart", ui.transitionToWaitingList);

// Notifies user when a crawler is done
pubsub.subscribe("crawlNotify", (currentCrawledObj) => {
  const list = JSON.parse(localStorage.getItem("list"));
  console.log(currentCrawledObj);
  const waitItems = Array.from(document.querySelectorAll(".wait-item"));
  waitItems.map((waitItem) => {
    const itemText = waitItem.children[0].textContent;
    if (itemText.includes(currentCrawledObj.URLSeed)) {
      if (currentCrawledObj.IsSuccess == true) {
        waitItem.dataset.state = "done";
        console.log(currentCrawledObj.Message);
      } else {
        waitItem.dataset.state = "error";
        const errorContainer = document.createElement("div");
        errorContainer.textContent = currentCrawledObj.Message + ": ";
        errorContainer.style.color = "#ed5e5e";
        waitItem.parentElement.insertBefore(errorContainer, waitItem);
      }

      // UPDATING LIST
      const updatedList = list.map((item) => {
        if (itemText === item.url) {
          return { url: item.url, state: waitItem.dataset.state };
        }
        return item;
      });
      localStorage.setItem("list", JSON.stringify(updatedList));
      // UPDATING LIST
    }
  });
});

pubsub.subscribe("crawlDone", () => {
  console.log("TEST: DONE CRAWLING");
  const newListBtn = document.getElementById("new-list-btn");
  newListBtn.style.display = "block";
  cookiesUtil.clearAllCookies();
  localStorage.setItem("list", "[]");
});
