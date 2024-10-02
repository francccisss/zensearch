const pageUrls = {
  home: "/",
  crawlProcess: "#crawl-section/process",
  results: "/search/results",
};

function showPage(path) {
  const pagesEl = document.querySelectorAll('main[id*="-page"]');
  pagesEl.forEach((page) => {
    page.hidden = true;
    page.style.display = "none";
  });
  if (path === pageUrls.home) {
    console.log(pageUrls.home);
    document.getElementById("search-page").hidden = false;
    document.getElementById("search-page").style.display = "flex";
  } else if (path === pageUrls.results) {
    console.log(pageurls.home);
    document.getElementById("results-page").hidden = false;
  } else {
    console.log(pageUrls.home);
    document.getElementById("process-page").hidden = false;
  }
}

export default { showPage };
