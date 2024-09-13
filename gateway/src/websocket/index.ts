import { WebSocketServer } from "ws";

const wss = new WebSocketServer();
const EVENTS = {
  search: "search",
};

wss.on(EVENTS.search, (s) => {});

export default wss;
