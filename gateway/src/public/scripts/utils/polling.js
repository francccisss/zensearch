import pubsub from "./pubsub.js";

async function poll(webUrls) {
  console.log("Poll something");
  try {
    setTimeout(() => {
      console.log(webUrls);
    }, 3 * 1000);
    const isSuccess = false;
    if (!isSuccess) {
      throw new Error("Error message from fetch");
    }

    // when polling when the data returned by crawler does not match
    // the length of the client's crawl list, poll and skip this function

    // If the crawler returns all of the list of urls the client
    // requested to crawl, then return this one
    pubsub.publish("crawlDone", {
      message: "Done",
      data: ["your", "unindexed", "list", "that", "were", "crawled"],
    });
    return {};
  } catch (err) {
    pubsub.publish("crawlError", {
      isPolling: false,
      message: "Polling",
      data: null,
    });
  }
}

async function loop() {
  await poll();
}

export default { poll };
