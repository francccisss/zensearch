const results_container = document.getElementById("search-results");

function search_item_component(item) {
  const { Title, Webpage_url, Contents } = item;
  const html_string = `<li class="searched-item">
    <a href="${Webpage_url}">${Title}</a>
    <span>${Contents}</span>
    </li>`;
  const parser = new DOMParser();
  return parser.parseFromString(html_string, "text/html").documentElement;
}

export default { search_item_component, results_container };
