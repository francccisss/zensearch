# zensearch
A small scale Distributed Search Engine using Golang for Crawling, indexing and ranking each document based on their relevancy to the user's search query.

## What does it do
A distributed search engine where user's are able to control what they can search, they can manually crawl specific websites of their liking and based on what they want to work with everyday. User's can crawl the web with a click of a button, and while crawling they can continue using the search feature to query existing webpages in their database, crawling might take some time because of security reasons and network throttling mechanism used by different website authors.

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
You will take full responsibility in the event that you will be blocked by a website author whose website you're crawling, so make sure you're crawling a website that would generally accept web crawlers and has a rate-limiting mechanism in their services, I have implemented a rudimentary rate-limiting mechanism in the crawler in `crawler/page_navigator.go` file called `requestDelay()`.


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
	min := 600
	max := 10000
	base := int(math.Log(float64(pn.mselapsed)))

	fmt.Printf("CURRENT ELAPSED TIME: %d\n", pn.mselapsed)
	if pn.interval < min {
		pn.interval = (pn.interval + base) * multiplier * 2
		fmt.Printf("INCREASE INTERVAL x2: %d\n", pn.interval)
	} else if pn.interval < max {
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

## How to Build

- Run this command to create an instance of rabbitmq Message broker.
```
docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:4.0-management
```
- Run the `build-dependencies.sh` script to install dependencies and build the services.
- Run each services separately

### Different branches
- If you want to run all of the services through docker please refer to the `deployment` branch where docker compose is used since each services heavily relies on rabbitmq for inter-process communication.

- The reason why this is a separate branch is because this branch hosts every service in the host machine through `amqp://<hostname>:/5672` to connect to the rabbitmq so the url would be `localhost` instead of the `rabbitmq`'s domain within the zensearch_network in docker compose.


### Chrome driver dependency
- For the crawler dependency `chromewebdriver` in the `test-environment` branch, the driver is within the same directory as the crawler service, but in the `deployment` branch, the `chromedriver` dependency should be in your PATH `/usr/local/bin` or `/usr/bin/` variable the same way it is set in the `Dockerfile` so please install the `chromewebdriver` dependency.

- You might notice that the string `chromeDriverPath=` variable in the `pkg/webdriver.go` file in the branch `test-environment` is set to the relative path but for the `deployment` branch since the `chromewebdriver` is in your PATH in we can just call `chromewebdriver`, so keep that in mind.


## How to Run
As mentioned earlier regarding different branches, use the code snippet above to create an instance of rabbitmq if you're not in `deployment` branch.

### Database
The project uses Sqlite3 database which is stored within `database-service/dist/website_collection.db`, you can go into it if you have `sqlite3` installed in your system and if not go ahead and install then after that:
- `cd` to the `database-service/dist/website_collection.db`
- run `sqlite3 website_collection.db`

There are three tables:
- `indexed_sites`
- `known_sites` (i know both of these mean the same thing)
- `webpages`

These are their schemas:
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

### Modifying data within sqlite3 in Docker Compose
I have not yet implemented a way for users to delete individual crawled websites, so in order for the users to remove users can navigate into the running container for `zensearch_db` or `zensearch_db-1`.

Issues when deleting might be necessary is when:
- crawler gets blocked so the crawler was not able to get all of the contents of that website.
 - recrawling would prompt the database service to return that the website has already been indexed because it saves the webpages up until the point where it was blocked which marks it as `known` or `indexed`
 - users will have to manually remove the indexed website in the database


#### How to modify a running database container
- run `docker exec -ti zensearch_db-1 sh` this command lets use create a terminal session within the running database container.
- `cd`into the `webiste_collection.db`
- run the same command for sqlite3 to modify the database


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
[Selenium](https://pkg.go.dev/github.com/tebeka/selenium)

#### Selenium Driver Dependencies (IMPORTANT)
[Chrome Driver Docs](https://developer.chrome.com/docs/chromedriver)
- Chrome Web Driver is within the `web-crawler-service/pkg/chrome` folder.
- This is needed for the client (crawler) to communicate with the web driver server via http and pass any api calls from the web driver server to the Web Devtools via web driver protocol. eg: `client (http)-> web driver (web driver protocol)-> devtools`


[Chrome Browser](https://www.google.com/chrome/)
[XFVB virtual frame buffer](https://www.x.org/releases/X11R7.6/doc/man/man1/Xvfb.1.xhtml)
- This is a system level dependency use whatever you have to install it.

#### Search Engine Service
[Go](https://go.dev/)
