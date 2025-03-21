import crawlInput from "./components/crawl_input/index.js";
import ui from "./ui/index.js";
import cookiesUtil from "./utils/cookies.js";
import pubsub from "./utils/pubsub.js";
import client from "./client_operations/index.js";
import clientws from "./client_operations/websocket.js";
import crawl_input from "./components/crawl_input/index.js";

const sidebar = document.getElementById("sidebar-container");
const searchBtn = document.getElementById("search-btn");
const openSbBtn = document.getElementById("add-entry-sb-btn");
const crawlBtn = document.querySelector(".crawl-btn");
const crawledData = new Map();
let isCrawling = false;

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

pubsub.subscribe("crawlReceiver", (msg) => {
  const { job_count } = cookiesUtil.extractCookies();
  const uint8 = new Uint8Array(msg.data_buffer.data);
  const decoder = new TextDecoder();
  const decodedBuffer = decoder.decode(uint8);
  const parseDecodedBuffer = JSON.parse(decodedBuffer);

  crawledData.set(parseDecodedBuffer.Url, parseDecodedBuffer);
  clientws.ackMessage();
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
    if (itemText.includes(currentCrawledObj.Url)) {
      if (currentCrawledObj.isSuccess) {
        waitItem.dataset.state = "done";
        console.log(currentCrawledObj.Message);
      } else {
        waitItem.dataset.state = "error";
        const errorContainer = document.createElement("div");
        errorContainer.textContent = currentCrawledObj.Message + ": ";
        errorContainer.style.color = "#ed5e5e";
        waitItem.parentElement.insertBefore(errorContainer, waitItem);
        console.log(currentCrawledObj.Message);
      }

      // UPDATING LIST
      const list = JSON.parse(localStorage.getItem("list"));
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

pubsub.subscribe("crawlDone", (currentCrawledObj) => {
  const newListBtn = document.getElementById("new-list-btn");
  newListBtn.style.display = "block";
  cookiesUtil.clearAllCookies();
  localStorage.clear();
});
