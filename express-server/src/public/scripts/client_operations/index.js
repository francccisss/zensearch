import cookiesUtil from "../utils/cookies.js";
import pubsub from "../utils/pubsub.js";
import clientws from "./websocket.js";
const ORIGIN = "http://localhost:8080";

// Upgrade http to websocket connection
async function checkListAndUpgrade(webUrls) {
  // try catch if an error while sending post request
  let responseObj = {};
  try {
    pubsub.publish("checkAndUpgradeStart");
    const sendWebUrls = await fetch(`${ORIGIN}/crawl`, {
      mode: "cors",
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(webUrls),
    });
    // specific for network errors
    if (sendWebUrls.ok === false) {
      throw new Error(sendWebUrls.statusText);
    }
    responseObj = await sendWebUrls.json();
    // For handling crawl list to be returned if
    // it has already been indexed, before upgrading
    // to a websocket connection.
    if (responseObj.is_crawling === false) {
      throw new Error(responseObj.message);
    }
    pubsub.publish("checkAndUpgradeDone", {});
    return responseObj.crawl_list;
  } catch (err) {
    pubsub.publish("checkAndUpgradeError", {
      status: "error",
      statusCode: responseObj.statusCode,
      message: err.message,
      data: responseObj.crawl_list,
    });
    throw new Error(err.message);
    return null;
  }
}
async function sendCrawlList() {
  const unhiddenInputs = document.querySelectorAll(
    'input.crawl-input:not([data-hidden="true"])',
  );
  const inputValues = Array.from(unhiddenInputs).map((input) => input.value);
  try {
    // checkListAndUpgrade returns the list else throws an error and returns null.
    const unindexed_list = await checkListAndUpgrade(inputValues);
    console.log("Transition to waiting area for crawled list.");

    const { message_type, job_id } = cookiesUtil.extractCookies();
    // On Successful database check for unindexed list, send the list to the
    // websocket server to start crawling the unindexed list.
    const message = { message_type, unindexed_list, meta: { job_id } };
    clientws.ws.send(JSON.stringify(message));
    // might return an error so we need to handle it before we transition
    // to waiting area.

    const mappedList = unindexed_list.map((item) => ({
      url: item,
      state: "loading",
    }));
    pubsub.publish("crawlStart", mappedList);
    // will be used for persistent ui for crawling state
    localStorage.setItem("list", JSON.stringify(mappedList));
  } catch (err) {
    console.error(err.message);
  }
}

export default { checkListAndUpgrade, sendCrawlList };
