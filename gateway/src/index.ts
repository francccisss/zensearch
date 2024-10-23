import http from "http";
import rabbitmq from "./rabbitmq/index";
import WebsocketService from "./websocket";
import { WebSocketServer } from "ws";
import app from "./express_server";

const PORT = 8080;
(async function start_server() {
  const http_server = http.createServer(app);
  /*
   Connect Rabbitmq
   Creates an indefinite loop to listen/receive
   new messages from the message broker.

   TODO Create a class for rabbitmq
  */
  const rbq_client = await rabbitmq.client.connectClient();

  // Connect Websocket for search results retrieved
  const wss: WebSocketServer = new WebSocketServer({ server: http_server });
  const ws_service = new WebsocketService(wss);
  ws_service.handler();

  /*
   Since actions from users eg: sending a crawl task,
   polling crawled tasks, and sending search queries are all handled
   by the express-server these rabbitmq handlers are specifically used for websockets

   The `init_websocket_channel_queues` asserts message queues.

   The `websocket_channel_listener` initializes listeners and takes in a cb
   function from the websocket server that pushes messages back to the client
   once the websocket_channel_listener consumes a message from the search engine  or Web crawler service.
  */
  await rbq_client.init_channel_queues();
  // ERROR this object undefined when used as callback function
  await rbq_client.websocket_channel_listener(
    ws_service.send_crawl_results_to_client.bind(ws_service),
  );
  await rbq_client.search_channel_listener();

  // Start HTTP server
  http_server.listen(PORT, () => {
    console.log("Listening to Port:", PORT);
  });
  /*
   Catching errors propogated by these initializers defined inside
   in the try block
  */
})().catch(console.error);
