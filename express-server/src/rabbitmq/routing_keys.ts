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
const EXPRESS_CRAWLER_QUEUE = "express_crawler_queue";
// callback queue passed by crawler to database, express uses it consume message
// from database after crawler successfully indexed and stored webpages in db
const DB_EXPRESS_INDEXING_CBQ = "db_express_indexing_cbq";

// SEARCH ENGINE ROUTING KEYS
// queue for sending search query to search engine from express server
const EXPRESS_SENGINE_QUERY_QUEUE = "express_sengine_query_queue";

// a callback queue to consume from search engine to express server
// after ranking webpages
const SENGINE_EXPRESS_QUERY_CBQ = "sengine_express_query_cbq";

// DB ROUTING KEYS
// route keys for checking db if the array of urls already exists
// or websites have already been indexed
const EXPRESS_DB_CHECK_QUEUE = "express_db_check_queue";
const DB_EXPRESS_CHECK_CBQ = "db_express_check_cbq";

export {
  EXPRESS_CRAWLER_QUEUE,
  EXPRESS_DB_CHECK_QUEUE,
  EXPRESS_SENGINE_QUERY_QUEUE,
  DB_EXPRESS_CHECK_CBQ,
  DB_EXPRESS_INDEXING_CBQ,
  SENGINE_EXPRESS_QUERY_CBQ,
};
