const ws = new WebSocket("ws://localhost:8080");

ws.addEventListener("open", (event) => {
  console.log("Connected");
});

ws.addEventListener("message", (event) => {
  console.log("Message from server: %s", event.data);
});

export default ws;
