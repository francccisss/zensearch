// I hate javascript

console.log("Initialized Script");
import polling from "../utils/job_polling.js";
import PubSub from "../utils/pubsub.js";

const crawl_btn = document.getElementById("crawl-btn");
const crawl_btn_container = document.getElementById("crawl-btn").parentElement;
const POLL_INTERVAL = 3;
const spinner = document.querySelector("span.process-spinner");
const search_bar = document.querySelector('input[type="search"]');
const crawler_rsp_cont = document.querySelector("#crawler-response-container");
const pubsub = new PubSub();
let is_crawling = false;

// to continue polling when users refresh the page

polling.init(
  // polling()
  () => {
    crawl_btn.disabled = true;
    crawl_btn_container.classList.add("polling");
    crawl_btn.textContent = "Crawling";
  },
  // done()
  (message) => {
    console.log(message);
    crawl_btn.textContent = "Crawl Websites";
    crawl_btn.disabled = false;
    crawl_btn_container.classList.remove("polling");
    search_bar.parentElement.style.display = "block";
  },
);

function update_crawl_list(response) {
  // Resets the list
  Array.from(crawler_rsp_cont.lastElementChild.children).forEach((li) => {
    li.remove();
  });
  crawler_rsp_cont.firstElementChild.textContent = response.message;
  response.crawl_list.forEach((item) => {
    const crawl_item = document.createElement("li");
    crawl_item.textContent = item;
    crawler_rsp_cont.lastElementChild.appendChild(crawl_item);
  });
}

async function handle_on_crawl(target) {
  try {
    console.log("crawl");
    const crawl = await fetch("http://localhost:8080/crawl", {
      method: "POST",
    });
    const response = await crawl.json();
    search_bar.parentElement.style.display = "none";
    if (!response.is_crawling) {
      update_crawl_list(response);
      return;
    }
    target.disabled = true;
    target.textContent = "Crawling";
    crawl_btn_container.classList.add("polling");
    const poll = await polling.loop();
    target.disabled = false;
    crawl_btn_container.classList.remove("polling");
    target.parentElement.style.display = "block";
    target.textContent = "Success";
  } catch (err) {
    crawl_btn.disabled = false;
    console.error(err.message);
  }
}

crawl_btn.addEventListener("click", async (e) => {
  const target = e.currentTarget;
  await handle_on_crawl(target);
});
