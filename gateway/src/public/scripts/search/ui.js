const results_container = document.getElementById("search-results");

function search_item_component(item) {
  const { Title, Webpage_url, Contents } = item;
  const html_string = `<li class="searched-item">
    <a href="${Webpage_url}">${Title}</a>
    <span>${Contents}</span>
    </li>`;
  const parser = new DOMParser();
  return parser
    .parseFromString(html_string, "text/html")
    .documentElement.querySelector("li.searched-item");
}

function render_webpages(webpages) {
  results_container.replaceChildren();
  webpages.forEach((page) => {
    results_container.append(search_item_component(page));
  });
}

export default { search_item_component, results_container, render_webpages };
