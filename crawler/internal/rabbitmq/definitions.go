package rabbitmq

type RabbitMQDefinitions struct {
	RBExchange    exchange    `yaml:"exchange"`
	RBRoutingKeys routingKeys `yaml:"routing_keys"`
	RBQueues      queues      `yaml:"queues"`
}

type exchange struct {
	General string `yaml:"general"`
	Crawler string `yaml:"crawler"`
}

type routingKeys struct {
	ExpressServerKeys struct {
		ES_CR_REQUEST string `yaml:"es_cr_request"`
	} `yaml:"express_server_keys"`
	CrawlerKeys struct {
		CR_DB_INDEXING string `yaml:"cr_db_indexing"`
		CR_DB_ENQUEUE  string `yaml:"cr_db_enqueue"`
		CR_DB_DEQUEUE  string `yaml:"cr_db_dequeue"`
		CR_DB_GETLEN   string `yaml:"cr_db_getlen"`
	} `yaml:"crawler_keys"`
}

type queues struct {
	ExpressServerQueues struct {
		ES_CR_REQUEST_QUEUE string `yaml:"es_cr_request_queue"`
		ES_CR_REQUEST_CBQ   string `yaml:"es_cr_request_cbq"`
	} `yaml:"express_server_queues"`

	CrawlerQueues struct {
		CR_DB_INDEXING_QUEUE string `yaml:"cr_db_indexing_queue"`
		CR_DB_INDEXING_CBQ   string `yaml:"cr_db_indexing_cbq"`

		CR_DB_ENQUEUE_QUEUE string `yaml:"cr_db_enqueue_queue"`
		CR_DB_ENQUEUE_CBQ   string `yaml:"cr_db_enqueue_cbq"`

		CR_DB_DEQUEUE_QUEUE string `yaml:"cr_db_dequeue_queue"`
		CR_DB_DEQUEUE_CBQ   string `yaml:"cr_db_dequeue_cbq"`

		CR_DB_GETLEN_QUEUE string `yaml:"cr_db_getlen_queue"`
		CR_DB_GETLEN_CBQ   string `yaml:"cr_db_getlen_cbq"`
	}
}
type CrawlerDefinitions struct {
	Exchange    exchange
	RoutingKeys RoutingKeys
	Queues      Queues
}

type RoutingKeys struct {
	ES_CR_REQUEST  string
	CR_DB_INDEXING string
	CR_DB_ENQUEUE  string
	CR_DB_DEQUEUE  string
	CR_DB_GETLEN   string
}
type Queues struct {
	ES_CR_REQUEST_QUEUE string
	ES_CR_REQUEST_CBQ   string

	CR_DB_INDEXING_QUEUE string
	CR_DB_INDEXING_CBQ   string

	CR_DB_ENQUEUE_QUEUE string
	CR_DB_ENQUEUE_CBQ   string

	CR_DB_DEQUEUE_QUEUE string
	CR_DB_DEQUEUE_CBQ   string

	CR_DB_GETLEN_QUEUE string
	CR_DB_GETLEN_CBQ   string
}
