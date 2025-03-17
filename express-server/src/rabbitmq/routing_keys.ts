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
const ES_CRAWLER_QUEUE = "es_crawler_queue";

// SEARCH ENGINE ROUTING KEYS
// queue for sending search query to search engine from express
const ES_SEARCH_QUEUE = "es_search_queue";
// a callback queue to consume from search engine to express
// after ranking webpages
const SEARCH_ES_CB = "search_es_cbq";

// DB ROUTING KEYS
// route keys for checking db if the array of urls already exists
// or websites have already been indexed
const ES_DB_CHECK_QUEUE = "es_db_check_queue";
const DB_ES_CHECK_CBQ = "db_es_check_cbq";
// callback queue passed by crawler to database, express uses it consume message
// from database after crawler successfully indexed and stored webpages in db
const DB_ES_SUCCESS_INDEXING_CBQ = "db_es_success_indexing_cbq";

export {
  ES_SEARCH_QUEUE,
  SEARCH_ES_CB,
  ES_CRAWLER_QUEUE,
  DB_ES_SUCCESS_INDEXING_CBQ,
  ES_DB_CHECK_QUEUE,
  DB_ES_CHECK_CBQ,
};
