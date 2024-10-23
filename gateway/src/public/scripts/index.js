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

searchBtn.addEventListener("click", async () => {
  await client.sendSearchQuery();
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
pubsub.subscribe("checkAndUpgradeError", ui.errorsui.handleCrawlErrors);

pubsub.subscribe("crawlReceiver", (msg) => {
  const { job_count } = cookiesUtil.extractCookies();
  const uint8 = new Uint8Array(msg.data_buffer.data);
  const decoder = new TextDecoder();
  const decodedBuffer = decoder.decode(uint8);
  const parseDecodedBuffer = JSON.parse(decodedBuffer);

  console.log(parseDecodedBuffer);
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
      if (currentCrawledObj.Message === "Success") {
        waitItem.dataset.state = "done";
      } else if (currentCrawledObj.Message === "Error") {
        waitItem.dataset.state = "error";
      }

      // UPDATING LIST
      const list = JSON.parse(localStorage.getItem("list"));
      const updatedList = list.map((item) => {
        console.log({ itemText, listItemText: item.url });
        if (itemText === item.url) {
          return { url: item.url, state: waitItem.dataset.state };
        }
        return item;
      });
      console.log(list);
      localStorage.setItem("list", JSON.stringify(updatedList));
      // UPDATING LIST
    }
  });
});

pubsub.subscribe("crawlDone", (currentCrawledObj) => {
  console.log(localStorage);
  const newListBtn = document.getElementById("new-list-btn");
  newListBtn.style.display = "block";
  cookiesUtil.clearAllCookies();
  localStorage.clear();
  console.log("Transition to SEARCH");
});
