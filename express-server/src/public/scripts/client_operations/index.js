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
    return null;
  }
}
async function sendCrawlList() {
  const unhiddenInputs = Array.from(
    document.querySelectorAll('input.crawl-input:not([data-hidden="true"])'),
  ).filter((input) => input.value !== "");
  const inputValues = unhiddenInputs.map((input) => input.value);
  let invalidList = [];
  try {
    if (inputValues.length === 0) {
      throw new Error(
        "Your crawl list is empty, please enter the websites you want to crawl: ",
      );
    }
    invalidList = checkURLList(inputValues);
    if (invalidList.length !== 0) {
      throw new Error(
        `Some of the items in this list are invalid URLs make sure to include a valid/supported protocol schema eg: "http","https" and TLD eg: ".com", ".dev"
        ".lol" etc.`,
      );
    }
    console.log("valid");
    // checkListAndUpgrade returns the list else throws an error and returns null.
    const unindexed_list = await checkListAndUpgrade(inputValues);
    // DONT HANDLE THE ERRORS OF CHECKLISTANDUPGRADE JUST
    if (unindexed_list === null) return;
    console.log("Transition to waiting area for crawled list.");

    const { message_type, job_id } = cookiesUtil.extractCookies();
    // On Successful database check for unindexed list, send the list to the
    // websocket server to start crawling the unindexed list.
    const message = { message_type, unindexed_list, meta: { job_id } };
    clientws.ws.send(JSON.stringify(message));
    // might return an error so we need to handle it before we transition
    // to waiting area.
    const waitingList = unindexed_list.map((item) => ({
      url: item,
      state: "loading",
    }));
    pubsub.publish("crawlStart", waitingList);
    // will be used for persistent ui for crawling state
    localStorage.setItem("list", JSON.stringify(waitingList));
  } catch (err) {
    console.error(err);
    pubsub.publish("checkAndUpgradeError", {
      status: "error",
      statusCode: 0,
      message: err.message,
      data: invalidList,
    });
  }
}

function checkURLList(list) {
  const invalidList = [];
  let message = "";
  for (let i = 0; i < list.length; i++) {
    try {
      if (!URL.canParse(list[i])) {
        throw new Error("Unable to parse url");
      }
      const setURL = new URL(list[i]);
      const validProtocol = ["http", "https"].includes(
        setURL.protocol.replace(":", ""),
      );
      if (!validProtocol) {
        throw new Error("Invalid/Unsupported protocol schema");
      }
    } catch (err) {
      console.error(err.message);
      invalidList.push(list[i]);
    }
  }
  return invalidList;
}

export default { checkListAndUpgrade, sendCrawlList };
