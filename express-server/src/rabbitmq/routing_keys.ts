/*
 * Message queue convention
 * <source>_<destination>_<function (optional)>_<queue|cbq (call back queue)>
 *
 * since each endpoint does not implement any fan-out message to multiple
 * services, its easier to isolate the destination in this naming convention
 *
 * that's just for me.
 */

const ES_CRAWLER_QUEUE = "es_crawler_queue";
const CRAWLER_ES_CBQ = "crawler_es_cbq";
const ES_SEARCH_QUEUE = "es_search_queue";
const SEARCH_ES_CB = "es_search_cbq";

// route keys for checking db if the array of urls already exists
// or websites have already been indexed
const ES_DB_CHECK_QUEUE = "es_db_check_queue";
const DB_ES_CHECK_CBQ = "db_es_check_cbq";

export {
  ES_SEARCH_QUEUE,
  SEARCH_ES_CB,
  ES_CRAWLER_QUEUE,
  CRAWLER_ES_CBQ,
  ES_DB_CHECK_QUEUE,
  DB_ES_CHECK_CBQ,
};
