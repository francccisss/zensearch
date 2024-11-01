# zensearch
A distributed search engine where user's are able to control what they can search, they can manually crawl specific websites of their liking and based on what they want to work with everyday. User's can crawl the web with a click of a button, and while crawling they can continue using the search feature to query existing webpages in their database, crawling might take some time because of security reasons and network throttling mechanism used by different website authors.


## Snapshots
![home-zensearch](https://github.com/user-attachments/assets/2851b18b-ea47-4e81-9e3b-9f9daa912b4c)
![crawl-section-zensearch](https://github.com/user-attachments/assets/8a968fa9-0326-4fc7-b6f2-3bff5236e909)
![crawling-sample](https://github.com/user-attachments/assets/2e57125c-af47-41d4-8724-300edaba2892)
![search-example](https://github.com/user-attachments/assets/a181e8e3-afbe-4b96-9570-185b9f14cf7f)


## Concepts
**TF-IDF**: "In information retrieval, tf–idf (also TF*IDF, TFIDF, TF–IDF, or Tf–idf), short for term frequency–inverse document frequency, is a measure of importance of a word to a document in a collection or corpus, adjusted for the fact that some words appear more frequently in general.Like the bag-of-words model, it models a document as a multiset of words, without word order. It is a refinement over the simple bag-of-words model, by allowing the weight of words to depend on the rest of the corpus." [source](https://en.wikipedia.org/wiki/Tf%E2%80%93idf).

- Despite it being fast to determine the ranking of each document relative to the user's search query, it disregards the context of the query as long as a document matches the terms in the search query eg: "dog bites man" or "man bites dog" does not matter in the context of this model.

**Document Length Normalization**: Is a technique which mitigates the length bias of a document whose context might not be relevant to the query of a user, because a document's length is greater than all other documents that contains the user's query, the *term frequency* of a longer document **might** be much more higher compared to all other documents if we disregard the concentration of a lengthy document to the term.

- The Document length normalization mitigates the length of a long document by dividing: currentDocLength / avgDocLength and controlled by `b` which controls the normalization of the document to determine the concentration of the term in that document. if the term frequency is proportionate to the document length then that means the current document is relevant to the query, else if the document is longer than average and is not proportionate to the term frequency then it is most likely no the main focus of the document.

**Beyond TF-IDF using BM25**: I am no expert but from my understanding of the BM25 Model is that it is an instance of the TF-IDF but with super powers, where the relevancy of a document is controlled by constants `k1` & `b` where `k1` controls the weight of a term frequency in a document or how much impact this term has throughout a document.

- In `k1` if the constant is set to a lower value, it saturates the term very quickly which diminishes the term frequency as the term grows and stops to a certain point but if it is set to a higher value eg: `k1 = 2` it will grow a bit slower up to a point where it begins slow down the rate as the term grows.

- The `b` controls the normalization of the length of the document relative to the term's relevancy or controls the concentration of the term in the document, if the term is sparse and is not mentioned enough in a long document and if `b` is high from 0-1, then long documents will be punished which means they will be scored lower, but if a document mentions the term more frequently and is more concetrated throughout the whole document, the document will be scored higher, using `0` normalization will render the document to only consider the term frequency and not consider if the document is relative to the term.


## TODO
- [ ] Save the most recently crawled webpage for continuation.
- [ ] Create cancellation for crawling but still save the indexed pages up to that point.
- [ ] One click to clear database.
- [ ] Let users delete a website from the sqlite database from the client-side.
- [ ] Documentation.
- [ ] Create a list of known websites (websites that have been indexed.) on the front-end

## IMPORTANT FOR USERS OF THIS PROJECT
You will take full responsibility in the event that you will be blocked by a website author, so make sure you're crawling a website that would generally accept web crawlers and has a rate-limiting mechanism in their services, I have implemented a rudimentary rate-limiter for the crawler in `web-crawler-service/page_navigator.go` file called `requestDelay()`.


```
/*
using elapsed time from start to end of request in milliseconds and compressing
it using log to smooth the values for increasing intervals for each requests
such that it doesnt grow too much when multiplying intervals.

multiplier values:
  - 0 ignores all intervals
  - 1 increases slowly but is still fast and might be blocked
  - 2 sweet middleground

The first check for pn.interval < min is hack i dont know what else to do.
*/
func (pn *PageNavigator) requestDelay(multiplier int) {
	max := 10000
	base := int(math.Log10(float64(pn.mselapsed)))

	fmt.Printf("CURRENT ELAPSED TIME: %d\n", pn.mselapsed)
	if pn.interval < max {
		pn.interval = (pn.interval + base) * multiplier
		fmt.Printf("INCREASE INTERVAL: %d\n", pn.interval)
	} else if pn.interval > max {
		fmt.Printf("RESET INTERVAL: %d\n", pn.interval)
		pn.interval = 0
	}
	time.Sleep(time.Duration(pn.interval * 1000000))
}
```

So be careful and read their `robots.txt` file from their website `https://<website-hostname>/robots.txt`.

## How to Run

```
# make sure you have docker and docker compose installed
# run these commands

cd path/to/zensearch/


docker compose up
```

### Database file size for rabbitmq
Rabbitmq has a limited message size it can transfer from one service to another, so if a user sends a query to the search engine, the search engine will then query the database to retrieve all of the indexed webpages, but if the database is > than what the rabbitmq can hold, then rabbitmq will throw an error due to the size of the message being sent. so for now please refrain from crawling too much, I'll try to build a sliding window algorithm such that it can scale despite how large the size of database.


### Database
The project uses Sqlite3 database which is stored within `database-service/dist/website_collection.db`, you can go into it if you have `sqlite3` installed in your system and if not go ahead and install then after that:
- `cd` to the `database-service/dist/website_collection.db`
- run `sqlite3 website_collection.db`

**TABLES**:
- `indexed_sites`
- `known_sites` (i know both of these mean the same thing)
- `webpages`

**SCHEMAS**:
```
known_sites (
    id INTEGER PRIMARY KEY,
    url TEXT NOT NULL,
    last_added INTEGER NOT NULL
);
indexed_sites (
    id INTEGER PRIMARY KEY,
    primary_url TEXT NOT NULL UNIQUE,
    last_indexed INTEGER NOT NULL
);
webpages (
    parent INTEGER REFERENCES indexed_sites(id),
    id INTEGER PRIMARY KEY,
    url TEXT NOT NULL ,
    title TEXT,
    contents TEXT
);
```

#### Modifying data within sqlite3 in Docker Compose
I have not yet implemented a way for users to delete individual crawled websites, so in order for the users to remove their indexed websites, users will have to navigate into the running container.

#### How to modify a running database container
- if you are in the zensearch directory do `docker exec -ti db sh` else find the docker running docker container prefixed with `zensearch-` using `docker ps` copy the container id of `zensearch-db` then do the same `docker exec -ti <id of container> sh` remove the brackets.
- run `sqlite3 dist/webiste_collection.db`
- remove the indexed website's data

# Tools and Dependencies

#### Message Broker
[RabbitMQ](https://www.rabbitmq.com/)

#### Frontend
[No Framework just vanilla yavascript](https://frontendmasters.com/blog/you-might-not-need-that-framework/)

#### Backend Server
[ExpressJs](http://expressjs.com/)
[NodeJS](https://nodejs.org/en)

#### Database Service
[NodeJS](https://nodejs.org/en)
[Sqlite3 for Nodejs](https://www.npmjs.com/package/sqlite3)
[Sqlite](https://www.sqlite.org/index.html)

#### Crawler Service
[Go](https://go.dev/)
[Selenium Container](https://hub.docker.com/r/selenium/standalone-chrome)

#### Selenium Driver Dependencies (IMPORTANT)
[Chrome Driver Docs](https://developer.chrome.com/docs/chromedriver)
- Chrome Web Driver is within the `web-crawler-service/pkg/chrome` folder.
- This is needed for the client (crawler) to communicate with the web driver server via http and pass any api calls from the web driver server to the Web Devtools via web driver protocol. eg: `client (http)-> web driver (web driver protocol)-> devtools`


[Chrome Browser](https://www.google.com/chrome/)
[XFVB virtual frame buffer](https://www.x.org/releases/X11R7.6/doc/man/man1/Xvfb.1.xhtml)
- This is a system level dependency use whatever you have to install it.

#### Search Engine Service
[Go](https://go.dev/)
