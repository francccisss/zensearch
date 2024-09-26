console.log("Initialized Script");
import polling from "../utils/job_polling.js";
import PubSub from "../utils/pubsub.js";

const crawl_btn = document.getElementById("crawl-btn");
const crawl_btn_container = document.getElementById("crawl-btn").parentElement;
const POLL_INTERVAL = 3;
const spinner = document.querySelector("span.process-spinner");
const search_bar = document.querySelector('input[type="search"]');
const pubsub = new PubSub();
let is_crawling = false;

pubsub.subscribe("polling_event", handle_on_crawl);

// to continue polling when users refresh the page

polling.init(
  // polling()
  () => {
    crawl_btn.disabled = true;
    crawl_btn_container.classList.add("polling");
    crawl_btn.textContent = "Crawling...";
  },
  // done()
  (message) => {
    console.log(message);
    crawl_btn.disabled = false;
    crawl_btn_container.classList.remove("polling");
  },
);

async function handle_on_crawl() {
  try {
    console.log("crawl");
    const crawl = await fetch("http://localhost:8080/crawl", {
      method: "POST",
    });
    //await polling.loop();
    crawl_btn.disabled = false;
    crawl_btn_container.classList.remove("polling");
  } catch (err) {
    crawl_btn.disabled = false;
    console.error(err.message);
  }
}

crawl_btn.addEventListener("click", async (e) => {
  const target = e.currentTarget;
  //target.disabled = true;
  //target.textContent = "Crawling...";
  //crawl_btn_container.classList.add("polling");
  pubsub.publish("polling_event");
});
