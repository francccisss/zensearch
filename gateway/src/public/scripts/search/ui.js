const results_container = document.getElementById("search-results");

function search_item_component(item) {
  const { Title, Url, Contents } = item;
  const url = new URL(Url);
  const paths = url.pathname.split("/");
  const path_segments = paths.join(" > ");
  const html_string = `
  <li class="searched-item">
    <div>
      <a href="${Url}">${Title}</a>
      <small>${url.hostname} ${path_segments}</small>
    </div>
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
    if (page.Contents == "") return;
    results_container.append(search_item_component(page));
  });
}

export default { search_item_component, results_container, render_webpages };
