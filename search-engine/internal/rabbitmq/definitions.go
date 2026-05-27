package rabbitmq

type RabbitMQDefinitions struct {
	Exchange    exchange    `yaml:"exchange"`
	RoutingKeys routingKeys `yaml:"routing_keys"`
	Queues      queues      `yaml:"queues"`
}
type SearchEngineDefinitions struct {
	Exchange    exchange
	RoutingKeys struct {
		SE_DB_REQUEST string
	}
	Queues struct {
		SE_DB_REQUEST_QUEUE string
		SE_DB_REQUEST_CBQ   string
		ES_SE_QUERY_QUEUE   string
	}
}

type exchange struct {
	General string `yaml:"general"`
	Crawler string `yaml:"crawler"`
}

type routingKeys struct {
	SearchEngineKeys struct {
		SE_DB_REQUEST string `yaml:"se_db_request"`
	} `yaml:"search_engine_keys"`
}

type queues struct {
	SearchEngineQueues struct {
		SE_DB_REQUEST_QUEUE string `yaml:"se_db_request_queue"`
		SE_DB_REQUEST_CBQ   string `yaml:"se_db_request_cbq"`
		ES_SE_QUERY_QUEUE   string `yaml:"se_db_request_queue"`
	} `yaml:"search_engine_queues"`
}
