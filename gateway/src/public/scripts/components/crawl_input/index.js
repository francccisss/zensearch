const template = document.getElementById("crawl-input-template");

function createComponent() {
  const container = document.createElement("div");
  container.append(template.content.cloneNode(true));
  container.classList.add("url-input");
  console.log(container);
  return container;
}

export default { createComponent };
