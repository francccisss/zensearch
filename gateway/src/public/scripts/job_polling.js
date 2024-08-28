import ui from "./ui.js";

async function get_current_job() {
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

async function poll_loop() {
  console.log("polling");
  let is_polling = true;
  while (is_polling) {
    console.log("poll");
    if (document.cookie === "") {
      is_polling = false;
      ui.set_btn_processing(
        document.querySelector(".process-spinner"),
        is_polling,
      );
      break;
    }
    ui.set_btn_processing(
      document.querySelector(".process-spinner"),
      is_polling,
    );

    // to block timeout
    await new Promise((resolved) => {
      setTimeout(async () => {
        await get_current_job();
        resolved("Next");
      }, 3 * 1000);
    });
  }
}
export default { poll_loop, get_current_job };
