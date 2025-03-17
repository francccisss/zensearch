/*
 * Message queue convention
 * <source>_<destination>_<function (optional)>_<queue|cbq (call back queue)>
 *
 * since each endpoint does not implement any fan-out message to multiple
 * services, its easier to isolate the destination in this naming convention
 *
 * that's just for me.
 */

// CRAWLER ROUTING KEYS
// queue used for when crawler is about to save webpages it crawled
const CRAWLER_DB_INDEXING_QUEUE = "crawler_db_indexing_queue";
// callback queue passed by crawler to database, express uses it consume message
// from database after crawler successfully indexed and stored webpages in db
const DB_EXPRESS_INDEXING_CBQ = "db_express_indexing_cbq";

// SEARCH ENGINE ROUTING KEYS
// routing key used by search engine service to request database for webpages.
const SENGINE_DB_REQUEST_QUEUE = "sengine_db_request_queue";
// routing key to reply back to the search engine service's callback queue.
const DB_SENGINE_REQUEST_CBQ = "db_sengine_request_cbq";

// EXPRESS SERVER ROUTING KEYS
// route keys for checking db if the array of urls already exists
// or websites have already been indexed
const EXPRESS_DB_CHECK_QUEUE = "express_db_check_queue";
const DB_EXPRESS_CHECK_CBQ = "db_express_check_cbq";

export {
  CRAWLER_DB_INDEXING_QUEUE,
  DB_EXPRESS_INDEXING_CBQ,
  SENGINE_DB_REQUEST_QUEUE,
  DB_SENGINE_REQUEST_CBQ,
  EXPRESS_DB_CHECK_QUEUE,
  DB_EXPRESS_CHECK_CBQ,
};
