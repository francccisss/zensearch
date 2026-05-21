/*
 * Message queue convention
 * <source>.<destination>.<function (optional)>.<queue|cbq (call back queue)>
 *
 * since each endpoint does not implement any fan-out message to multiple
 * services, its easier to isolate the destination in this naming convention
 *
 * that's just for me.
 */

// CRAWLER ROUTING KEYS
// queue used for when crawler is about to save webpages it crawled
const CRAWLER_DB_INDEXING_QUEUE = "crawler.db.indexing.queue";
const DB_CRAWLER_INDEXING_CBQ = "db.crawler.indexing.cbq"

// SEARCH ENGINE ROUTING KEYS
// routing key used by search engine service to request database for webpages.
const SENGINE_DB_REQUEST_QUEUE = "sengine.db.request.queue";
// routing key to reply back to the search engine service's callback queue.
const DB_SENGINE_REQUEST_CBQ = "db.sengine.request.cbq";

// EXPRESS SERVER ROUTING KEYS
// route keys for checking db if the array of urls already exists
// or websites have already been indexed
const EXPRESS_DB_CHECK_QUEUE = "express.db.check.queue";
const DB_EXPRESS_CHECK_CBQ = "db.express.check.cbq";


// Frontier Queue Routing keys
const CRAWLER_DB_STORE_FRONTIER_QUEUE = "crawler.db.store.frontier.queue"
const CRAWLER_DB_CLEAR_FRONTIER_QUEUE = "crawler.db.clear.frontier.queue"

const CRAWLER_DB_DEQUEUE_FRONTIER_QUEUE = "crawler.db.dequeue.frontier.queue"
const DB_CRAWLER_DEQUEUE_FRONTIER_CBQ = "db.crawler.dequeue.frontier.cbq"

const CRAWLER_DB_LEN_FRONTIER_QUEUE = "crawler.db.len.frontier.queue"
const DB_CRAWLER_LEN_FRONTIER_CBQ = "db.crawler.len.frontier.cbq"


export {
	CRAWLER_DB_INDEXING_QUEUE,
	DB_CRAWLER_INDEXING_CBQ,

	SENGINE_DB_REQUEST_QUEUE,
	DB_SENGINE_REQUEST_CBQ,

	EXPRESS_DB_CHECK_QUEUE,
	DB_EXPRESS_CHECK_CBQ,


	CRAWLER_DB_STORE_FRONTIER_QUEUE,
	CRAWLER_DB_CLEAR_FRONTIER_QUEUE,
	CRAWLER_DB_DEQUEUE_FRONTIER_QUEUE,
	DB_CRAWLER_DEQUEUE_FRONTIER_CBQ,
	CRAWLER_DB_LEN_FRONTIER_QUEUE,
	DB_CRAWLER_LEN_FRONTIER_CBQ
}
