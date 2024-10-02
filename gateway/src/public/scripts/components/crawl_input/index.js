class Component extends HTMLElement {
  input;
  constructor() {
    super();
    this.#createComponent();
  }
  #createComponent() {
    const shadow = this.attachShadow({ mode: "open" });
    const temp = document.getElementById("crawl-input-temp");
    shadow.append(temp.content.cloneNode(true));
    this.classList.add("crawl-input");
  }
}
customElements.define("crawl-input", Component);

export default { Component };
