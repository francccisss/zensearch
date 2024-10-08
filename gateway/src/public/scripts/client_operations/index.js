import pubsub from "../utils/pubsub.js";

// Upgrade http to websocket connection
async function sendCrawlRequest(webUrls) {
  // try catch if an error while sending post request
  let responseObj = {};
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
  } catch (err) {
    pubsub.publish("crawlError", {
      status: "error",
      statusCode: responseObj.statusCode,
      message: err.message,
      data: responseObj.crawl_list,
    });
    throw new Error(err.message);
  }
}

export default { sendCrawlRequest };
