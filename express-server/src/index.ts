import http from "http";
import rabbitmq from "./rabbitmq/index.js";
import type { ExpressServerDefinition, RabbitMQDefinitions } from "./rabbitmq/rbq_definitions.js"
import WebsocketService from "./websocket/index.js";
import { WebSocketServer } from "ws";
import app from "./express_server/index.js";
import { exit } from "process";
import yaml from "js-yaml"
import fs from "fs"
import path from "node:path";


const httpServer = http.createServer(app);
const PORT = 8080;
await (async function start_server() {
	console.log(import.meta.dirname)
	const yamlfile = fs.readFileSync(path.resolve(import.meta.dirname, "../../rabbitmq.yml"), "utf8")
	// const yamlfile = "../../rabbitmq.yml"
	const doc = yaml.load(yamlfile, {}) as unknown as RabbitMQDefinitions
	const expressDef: ExpressServerDefinition = {
		exchange: {
			...doc.exchange
		}
		,
		queues: { ...doc.queues["express_server_queues"] as any },
		routing_keys: { ...doc.routing_keys["express_server_keys"] as any },

	}
	console.log(expressDef.exchange)
	console.log(expressDef.queues)
	console.log(expressDef.routing_keys)
	return
	console.log("Starting express server");

	httpServer.on("error", (err: any) => {
		if (err.code == "EADDRINUSE") {
			console.error(err.message);
			httpServer.close();
			console.log("exiting process");
			exit(1);
		}
	});
	/*
	 Connect Rabbitmq
	 Creates an indefinite loop to listen/receive
	 new messages from the message broker.
	*/
	const rbqConn = await rabbitmq.client.establishConnection(7);

	// Connect Websocket for search results retrieved
	const wss: WebSocketServer = new WebSocketServer({ server: httpServer });
	const wsService = new WebsocketService(wss);
	wsService.handler();

	// rbq_client.segmentGenerator();
	await rbqConn.initChannelQueues();

	rbqConn.crawlChannelListener(
		wsService.sendCrawlResultsToClient.bind(wsService),
	);
	(async () => {
		let segmentsReceived = 0;
		rbqConn.eventEmitter.on("newSegment", () => {
			segmentsReceived++;
		});

		rbqConn.eventEmitter.on("done", () => {
			console.log("Segments Received: %d", segmentsReceived);
			segmentsReceived = 0;
		});
	})();

	rbqConn.addSegmentsToQueue();
	rbqConn.searchChannelListener();

	// Start HTTP server
	httpServer.listen(PORT, () => {
		console.log("Listening to Port:", PORT);
	});
	/*
	 Catching errors propogated by these initializers defined inside
	 in the try block
	*/
})().catch((e: any) => {
	console.error(e);
	httpServer.close();
	console.log("SHUTDOWN");
	exit(1);
});
