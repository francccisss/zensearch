const ws = new WebSocket("ws://localhost:8080");

ws.addEventListener("open", (event) => {
  console.log("Connected");
});

ws.addEventListener("message", (event) => {
  const parsed_webpages = JSON.parse(event.data);
  console.log("Message from server: ");
  console.log(parsed_webpages);
});

export default ws;
