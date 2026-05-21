/*
 * Message queue convention
 * <source>_<destination>_<function (optional can be chained for more context at the cost of verbosity)>_<queue|cbq (callback queue)>
 *
 * since each endpoint does not implement any fan-out message to multiple
 * services, its easier to isolate the destination in this naming convention
 *
 * that's just for me.
 */

// CRAWLER ROUTING KEYS
// queue for requesting a crawl from express to crawler
const EXPRESS_CRAWLER_CRAWL_QUEUE = "express.crawler.crawl.queue"
const CRAWLER_EXPRESS_CRAWL_CBQ = "crawler.express.crawl.cbq"

// SEARCH ENGINE ROUTING KEYS
// queue for sending search query to search engine from express server
const EXPRESS_SENGINE_QUERY_QUEUE = "express.sengine.query.queue";

// a callback queue to consume from search engine to express server
// after ranking webpages
const SENGINE_EXPRESS_QUERY_CBQ = "sengine.express.query.cbq";

// DB ROUTING KEYS
// route keys for checking db if the array of urls already exists
// or websites have already been indexed
const EXPRESS_DB_CHECK_QUEUE = "express.db.check.queue";
const DB_EXPRESS_CHECK_CBQ = "db.express.check.cbq";

export {
	EXPRESS_CRAWLER_CRAWL_QUEUE,
	CRAWLER_EXPRESS_CRAWL_CBQ,
	EXPRESS_DB_CHECK_QUEUE,
	EXPRESS_SENGINE_QUERY_QUEUE,
	DB_EXPRESS_CHECK_CBQ,
	SENGINE_EXPRESS_QUERY_CBQ,
};
