import extract_cookies from "./extract_cookies.js";

async function poll_job() {
  console.log("Poll Job");
  let cookies = extract_cookies();
  const polling = await fetch(
    `http://localhost:8080/job?job_id=${cookies.job_id}&job_queue=${cookies.job_queue}&job_count=${cookies.job_count}`,
  );

  const polling_response = polling.ok ? await polling.json() : null;
  console.log(polling_response);
  return polling_response;
}

// need a timer to stop looping if the server might be down
async function loop() {
  console.log("polling");
  let is_polling = true;
  while (is_polling) {
    if (document.cookie === "") {
      is_polling = false;
      break;
    }

    const job = await poll_job();
    if (job.done) {
      return job;
    }
    // to block timeout
    await new Promise((resolved) => {
      setTimeout(async () => {
        resolved("Next");
      }, 3 * 1000);
    });
  }
}

async function init(polling, done) {
  if (document.cookie !== "") {
    polling();
  }
  const job = await loop();
  done(job);
}

export default { loop, poll_job, init };
