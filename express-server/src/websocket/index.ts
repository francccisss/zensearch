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
        if (data.toString() === "ACK") {
          // Client cant send a pong to the server so, server will have to handle that
          // ignore this
          console.log("Client sent ACK to websocket server.");
          return;
        }
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
    A callback function that is called within the channel listeners for consuming
    messages from the message queues, it handles the delivery of acknoledgments to
    the rabbitmq message broker based on the connection of the client through
    the websocket connection.

    Need to create a reliable data transfer pipeline for sending acks from client
    that it received the message from the websocket server
    such that the websocket server will be able to send an acknoledgment back
    to the message broker that the message from the message queue
    has been successfully acknowledged


    I just realized theres no way for me to reconnect the client
    using the same ephemeral port that is stored as a client object
    in the server's memory once the client disconnects from the server
    so this will just be requeued indefinitely. Wow

    Implement a websocket user authentication for client reconnection?
  */

  async sendCrawlResultsToClient(
    chan: Channel,
    msg: ConsumeMessage,
    message_type: string,
  ) {
    this.wss.clients.forEach((ws) => {
      let isAck = { state: false };
      console.log("Sending start");
      /* Handling Disconnected clients

         The requeuing of the message back to the message queue
         will call this function again.

         Need to somehow set a timer for on pong event, if the server has not received
         an acknowledgment from the alloted time, then we retransmit the message from message queue

         When nack is called on message queue channel, the message will be requeued.
        */

      /*
         Error code 406 from rabbitmq because of an Unkown delivery tag
         after we ack the last message, this is called again, right after acknowledging
         the previous one. which is why it throws a 406 error code.


         --LOGS--

         LOG: Message received from crawling
         Sending start
         {"Message":"Success","Url":"fzaid.vercel.app","WebpageCount":5}
         Client sent ACK
         Server received ACK
         Called from settimeout
         LOG: Message received from crawling
         Sending start
         {"Message":"Success","Url":"m7mad.dev","WebpageCount":6}
         Client sent ACK
         Server received ACK
         Server received ACK <-- Problem
        */

      /*
         Client does not need to send a "NACK" message since we can rely on tcp for currupted bits
         handling, but instead we will only handle lossy channel between client and server to
         retransmit messages from the message queue when client suddenly disconnects.

         BRUH after creating a listener for the previous message
         the previous listener still exists, ALWAYS REMOVE AN EVENT LISTENER AFTER USING IT,
         TO PREVENT DUPLICATE CALLS.
        */
      const messageHandler = function (message: RawData) {
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
            ws.off("message", messageHandler);
          }
        }
      };
      ws.on("message", messageHandler);

      const responseData = JSON.stringify({
        data_buffer: msg.content,
        message_type,
      });
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
          ws.off("message", messageHandler);
        }
      }, 3 * 1000);
    });
  }
}

export default WebsocketService;
