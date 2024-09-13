const ws = new WebSocket("ws://localhost:8080");

// init connection

ws.addEventListener("open", (event) => {
  console.log("Connected");
  ws.send("Hello Server!");
});

export default ws;
