package rabbitmq

// CRAWLER ROUTING KEYS
// queue used for when crawler is about to save webpages it crawled
const CRAWLER_DB_INDEXING_QUEUE = "crawler_db_indexing_queue"

// callback queue passed by crawler to database, express uses it consume message
// from database after crawler successfully indexed and stored webpages in db
const DB_EXPRESS_INDEXING_CBQ = "db_express_indexing_cbq"
