async function poll_job() {
  console.log("Poll Job");
  const [job_id, job_queue] = document.cookie.split("; ").map((c) => {
    const entries = c.split("=");
    const cookie_obj = { [entries[0]]: entries[1] };
    return cookie_obj;
  });
  const polling = await fetch(
    `http://localhost:8080/job?job_id=${job_id.job_id}&job_queue=${job_queue.job_queue}`,
  );
  const polling_response = await polling.text();
  console.log(polling_response);
}

async function loop() {
  console.log("polling");
  let is_polling = true;
  while (is_polling) {
    if (document.cookie === "") {
      is_polling = false;
      break;
    }
    // to block timeout
    await new Promise((resolved) => {
      setTimeout(async () => {
        await poll_job();
        resolved("Next");
      }, 3 * 1000);
    });
  }
}

async function init(polling, done) {
  if (document.cookie !== "") {
    polling();
  }
  await loop();
  done();
}

export default { loop, poll_job, init };
