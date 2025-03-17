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
const CRAWLER_DB_INDEXING_NOTIF_QUEUE = "crawler_db_notif_indexing_queue";

// a callback queue for notifying crawler that storing indexed webpages
// was successful or failure
const DB_CRAWLER_INDEXING_NOTIF_CBQ = "db_crawler_indexing_notif_cbq";

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
  CRAWLER_DB_INDEXING_NOTIF_QUEUE,
  DB_CRAWLER_INDEXING_NOTIF_CBQ,
  SENGINE_DB_REQUEST_QUEUE,
  DB_SENGINE_REQUEST_CBQ,
  EXPRESS_DB_CHECK_QUEUE,
  DB_EXPRESS_CHECK_CBQ,
};
