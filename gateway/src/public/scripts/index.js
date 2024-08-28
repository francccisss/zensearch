console.log("Initialized Script");
import polling from "./job_polling.js";
import PubSub from "./pubsub.js";

const crawl_btn = document.getElementById("crawl-btn");
const POLL_INTERVAL = 3;
const spinner = document.querySelector("span.process-spinner");
const pubsub = new PubSub();
let is_crawling = false;

pubsub.subscribe("polling_event", handle_on_crawl);

(async () => {
  if (document.cookie !== "") {
    crawl_btn.disabled = true;
    crawl_btn.textContent = "Crawling...";
  }
  await polling.poll_loop();
  crawl_btn.disabled = false;
})();

async function handle_on_crawl() {
  try {
    console.log("crawl");
    const crawl = await fetch("http://localhost:8080/crawl", {
      method: "POST",
    });
    await polling.poll_loop();
    crawl_btn.disabled = false;
  } catch (err) {
    console.error(err.message);
  }
}

crawl_btn.addEventListener("click", async (e) => {
  const target = e.currentTarget;
  target.disabled = true;
  target.textContent = "Crawling...";
  pubsub.publish("polling_event");
});
