console.log("Initialized Script");

const crawl_btn = document.getElementById("crawl-btn");

crawl_btn.addEventListener("click", (e) => {
  const target = e.currentTarget;
  target.textContent = "This might take a while...";
  target.disabled = true;
});
