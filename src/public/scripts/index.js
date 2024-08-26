console.log("Initialized Script");

const crawl_btn = document.getElementById("crawl-btn");
let is_crawling = false;

crawl_btn.addEventListener("click", (e) => {
  const target = e.currentTarget;
  target.textContent = "This might take a while...";
  target.disabled = true;
  is_crawling = true;
});
