import navigation from "./client_navigation/navigation.js";
import crawlInput from "./components/crawl_input/index.js";
import ui from "./ui/index.js";
import pubsub from "./utils/pubsub.js";

const sidebar = document.getElementById("crawl-list-sb");
const closeSbBtn = document.getElementById("close-sb");
const openSbBtn = document.getElementById("add-entry-sb-btn");
const listContainer = document.querySelector(".list-container");

window.addEventListener("load", () => {
  ui.init();
  navigation.showPage("/");
});

openSbBtn.addEventListener("click", () => {
  sidebar.classList.replace("inactive-sb", "active-sb");
});

sidebar.addEventListener("click", ui.sidebarActions);
pubsub.subscribe("removeEntry", crawlInput.updateEntries);
