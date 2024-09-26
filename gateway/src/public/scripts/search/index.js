import polling from "../utils/job_polling.js";
import client_websocket from "./websocket.js";
import extract_cookies from "../utils/extract_cookies.js";
import searchui from "./ui.js";

const search_input = document.getElementById("search-input");
const form = document.querySelector("form");
const cookies = extract_cookies();

// This is used after the client has Established/Upgraded a websocket connection with the server
form.addEventListener("submit", (e) => {
  e.preventDefault();
  const input_value = search_input.value;
  client_websocket.send(
    JSON.stringify({ q: input_value, job_id: cookies.job_id }),
  );
  console.log("Search query: %s", input_value);
  console.log("Fetching searched items");
});
