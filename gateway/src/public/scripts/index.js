import navigation from "./client_navigation/navigation.js";
import crawlInput from "./components/crawl_input/index.js";
import ui from "./ui/index.js";
import extract_cookies from "./utils/extract_cookies.js";
import pubsub from "./utils/pubsub.js";
import client from "./client_operations/index.js";
import ws from "./client_operations/websocket.js";

const sidebar = document.getElementById("crawl-list-sb");
const openSbBtn = document.getElementById("add-entry-sb-btn");
const crawlBtn = document.querySelector(".crawl-btn");
const crawledData = new Map();
// TODO Add documentations
// TODO Attach loop poll if user's refreshes the browser.

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
    // checkListAndUpgrade returns the list else throws an error and returns null.
    const unindexed_list = await client.checkListAndUpgrade(inputValues);
    console.log("Transition to waiting area for crawled list.");

    const { message_type, job_id } = extract_cookies();
    // On Successful database check for unindexed list, send the list to the
    // websocket server to start crawling the unindexed list.
    const message = { message_type, unindexed_list, meta: { job_id } };
    ws.send(JSON.stringify(message));
  } catch (err) {
    console.error(err.message);
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
pubsub.subscribe("checkAndUpgradeError", ui.errorsui.handleCrawlErrors);

pubsub.subscribe("crawlReceiver", (msg) => {
  const { job_count } = extract_cookies();
  const uint8 = new Uint8Array(msg.data_buffer.data);
  const decoder = new TextDecoder();
  const decodedBuffer = decoder.decode(uint8);
  const parseDecodedBuffer = JSON.parse(decodedBuffer);

  console.log(parseDecodedBuffer);
  crawledData.set(parseDecodedBuffer.url, parseDecodedBuffer);
  if (crawledData.size < job_count) {
    console.log("Less than");
    // Need to update the ui of the current crawled website
    // in the unindexed list.
    return;
  }
  console.log("Transition to search.");
  // if size is === to job_count transition page to search.
  // console.log(data);
});

const d = {
  data_buffer: {
    type: "Buffer",
    data: [
      123, 34, 109, 101, 115, 115, 97, 103, 101, 34, 58, 34, 83, 117, 99, 99,
      101, 115, 115, 34, 44, 34, 117, 114, 108, 34, 58, 34, 102, 122, 97, 105,
      100, 46, 118, 101, 114, 99, 101, 108, 46, 97, 112, 112, 34, 44, 34, 119,
      101, 98, 112, 97, 103, 101, 95, 99, 111, 117, 110, 116, 34, 58, 53, 125,
    ],
  },
  message_type: "crawling",
};
pubsub.publish("crawlReceiver", d);
