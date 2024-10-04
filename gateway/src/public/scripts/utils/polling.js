import extract_cookies from "./extract_cookies.js";
import pubsub from "./pubsub.js";

const pollUrl = "http://localhost:8080/job";

async function poll(webUrls) {
  console.log("Poll something");
  const { job_count, job_id, job_queue } = extract_cookies();
  let responseObj = {};
  try {
    const pollData = await fetch(
      `${pollUrl}?job_count=${job_count}&job_id=${job_id}&job_queue=${job_queue}`,
    );
    responseObj = { ...(await pollData.json()) };
    console.log(responseObj);

    if (pollData.ok === false) {
      throw new Error(pollData.statusText);
    }

    // for looping
    if (responseObj.done === false) {
      return responseObj.done;
    }

    pubsub.publish("pollingDone", {
      message: "Done",
      isPolling: false,
      data: responseObj.data,
    });
    return responseObj.done;
  } catch (err) {
    console.error(err);
    pubsub.publish("crawlError", {
      isPolling: false,
      message: err.message,
      data: null,
    });
  }
}

async function loop() {
  const { job_count } = extract_cookies();
  let isDone = false;
  while (!isDone) {
    isDone = await poll();
    if (isDone) {
      console.log("Stop polling");
      return;
    }
    await new Promise((resolve, reject) => {
      setTimeout(() => {
        console.log("Poll loop triggered.");
        resolve("");
      }, 2 * 1000);
    });
  }
}

export default { poll, loop };
