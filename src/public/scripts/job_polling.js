async function poll_current_job() {
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
  const crawl_btn = (document.getElementById("crawl-btn").textContent =
    polling_response);
}

export default poll_current_job;
