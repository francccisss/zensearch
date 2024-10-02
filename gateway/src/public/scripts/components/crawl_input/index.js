const template = document.getElementById("crawl-input-template");
const listContainer = document.querySelector(".crawl-list-container")
  .children[0];

function createComponent() {
  const container = document.createElement("div");
  container.append(template.content.cloneNode(true));
  container.classList.add("url-input");
  console.log(container);
  return container;
}

function addNewEntry() {
  listContainer.appendChild(createComponent());
}

function removeEntry(id) {
  const children = Array.from(listContainer.children);
  listContainer.replaceChildren(
    children.filter((child) => {
      if (child.id !== id) {
        return child;
      }
    }),
  );
}

export default { createComponent, addNewEntry };
