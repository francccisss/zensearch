import { WebSocketServer, WebSocket, RawData, Data } from "ws";
import rabbitmq from "../rabbitmq";
import { Channel, ConsumeMessage } from "amqplib";
import { EXPRESS_CRAWLER_QUEUE } from "../rabbitmq/routing_keys";

const EVENTS = {
  message: "message",
  connection: "connection",
  close: "close",
};

class WebsocketService {
  wss: WebSocketServer;
  isAlive: boolean = false;
  constructor(WS: WebSocketServer) {
    this.wss = WS;
  }

  /*
    Listens to new tcp connections using websocket protocol and receives/listens
    to new client messages that are pushed to the websocket server.
  */

  // Handler needs to mutliplex incoming messages from the client
  // if it a crawl request or a search request.

  async handler() {
    this.wss.on(EVENTS.connection, async (ws: WebSocket) => {
      console.log("connected");
      this.isAlive = true;

      ws.on(EVENTS.message, async (data: Data) => {
        // client sends an ack when a message from crawler was received
        if (data.toString() === "ACK") {
          // Client cant send a pong to the server so, server will have to handle that
          // ignore this
          console.log("Client sent ACK to websocket server.");
          return;
        }

        // After server check the user defined crawl list, client
        // will trigger a remote procedure to the websocket server by calling
        // `rabbitmq.client.crawl` along with the list to be crawled by the crawler

        // this is confusing because data could be anything
        // the first pass is used when data is just a pure string
        // which is used by the `sendCrawlResultsToClient` handler
        // for acknowledging notifications from crawler
        console.log("Message received");
        const decodeBuffer: {
          message_type: "crawling" | "searching";
          meta: { job_id: string };
          unindexed_list?: Array<string>;
          [key: string]: any;
        } = JSON.parse(data.toString());

        if (decodeBuffer.message_type === "crawling") {
          const serializeList = Buffer.from(
            JSON.stringify({ Docs: decodeBuffer.unindexed_list! }),
          );
          const success = await rabbitmq.client.crawl(serializeList, {
            queue: EXPRESS_CRAWLER_QUEUE,
            id: decodeBuffer.meta.job_id,
          });
          if (!success) {
            throw new Error(
              "Unable to send crawl list to web crawler service.",
            );
          }
        }
      });

      ws.on("close", () => {
        this.isAlive = false;
        console.log("Connection closed is alive is false");
      });
    });
  }

  /*
    A callback function that is called within the channel listeners when there is 
    a new message in the `EXPRESS_CRAWLER_QUEUE` queue, it handles the delivery of acknoledgments to
    the rabbitmq message broker based on the connection of the client through
    the websocket connection.
    (basically messages will only be acknowledge when client is connected)

    sendCrawlResultsToClient is called, creates a clientAckListener and sends the message from crawler to the client
    which then the client sends back an ack for the current message (the handler above, when on "message" is sent by 
    client it also receives the ack message so ignore that), the clientAckListener checks if message is ack
    (it should only just be ack, this function should be the one to call nack if the client has not received it in some
    alloted time and will requeue it).

    after that the listener is removed


  */

  async sendCrawlResultsToClient(
    chan: Channel,
    msg: ConsumeMessage,
    message_type: string,
  ) {
    this.wss.clients.forEach((ws) => {
      let isAck = { state: false };
      console.log("Sending start");

      const clientAckListener = function (message: RawData) {
        const ack: "ACK" | "NACK" = message.toString() as "ACK" | "NACK";
        if (ack === "ACK") {
          console.log("Server received ACK");
          isAck.state = true;
          try {
            if (isAck.state === true) {
              console.log("LOG: Acknowledge message from message queue");
              chan.ack(msg);
            }
          } catch (err) {
            const error = err as Error;
            console.log(
              "LOG: An error has occured while acknowledging sent message",
            );
            console.error(error.name);
            console.error(error.message);
          } finally {
            ws.removeListener("message", clientAckListener);
          }
        }
      };
      // listener handles message from the client when client sends an ack for the current
      // message from the crawler.
      ws.on("message", clientAckListener);

      const responseData = JSON.stringify({
        data_buffer: msg.content,
        message_type,
      });

      // send the crawl message status to client from the crawler
      ws.send(responseData, (err) => {
        if (err) {
          // This Nack will be labeled as a server | client error.
          // Still dont know how to handle this
          if (msg !== null) {
            chan.nack(msg, true, false);
          }
          console.error(err);
          throw new Error("ERROR: Unable to send message to client.");
        }
      });

      /*
       ## Requeuing when a client is ready to receive message from crawler
       Start timer to determine if connection is closed and is unable
       to send an Ack to the websocket sender if client disconnects while
       sending this message.

       This is only for the current message to be labeled as nack and then requeued.
       */
      setTimeout(() => {
        if (isAck.state === false) {
          if (msg === null) {
            console.log(
              "LOG: No Message to be Timedout: wait for new message.",
            );
            return;
          }
          console.log("LOG: Timeout, retransmit message.");
          chan.nack(msg, false, true);
          ws.removeListener("message", clientAckListener);
        }
      }, 3 * 1000);
    });
  }
}

export default WebsocketService;
