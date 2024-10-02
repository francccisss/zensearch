const sheet = new CSSStyleSheet();
sheet.replaceSync(`
  button,input{
    appearance: none;
    border-radius: 9999px;
    border:none;
    padding: var(--crawl-input-btn-padding);
    font-family:"Poppins";
  }
  button{
    position: absolute;
    left:100%;
    top:0%;
    background-color: var(--black);
    color: var(--light-text);
  }
  input{
    width:100%;
    background-color: var(--input-color);
  }
`);

class Component extends HTMLElement {
  input;
  constructor() {
    super();
    this.#createComponent();
  }
  #createComponent() {
    const shadow = this.attachShadow({ mode: "open" });
    const temp = document.getElementById("crawl-input-temp");
    shadow.adoptedStyleSheets = [sheet];
    shadow.append(temp.content.cloneNode(true));
    this.classList.add("crawl-input");
    this.style.position = "relative";
    //this.style.width = "70%";
  }
}
customElements.define("crawl-input", Component);

export default { Component };
