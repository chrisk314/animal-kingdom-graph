# Animal Kingdom Graph
This project contains a Wikipedia crawler implemented in Golang which builds a graph of 
the animal kingdom. The Colly web crawler package and Goquery package are used to 
traverse Wikipedia pages related to animal species. Data from each page related
to an animal species is stored in an Arango graph database along with the related 
taxonomic hierarchy.

There are many animal species which do not fit neatly into a consistent taxonomic heirarchy. 
To simplify things, only species whose taxonomic heirarchy matches the below sequence are 
extracted and stored.

kingdom -> phylum -> class -> order -> family -> genus -> species

## Install
To install the dependencies and build the binary run the below commands from the [wiki-scraper](./wiki-scraper/) directory.
```shell
go get
go build
```

## Run scraper
To run an Arango DB server run the below command.
```shell
docker-compose up
```
Now the scraper can be launched with the below command.
```shell
cd ./wiki-scraper && wiki_scraper
```

## Run visualiser
To run an Arango DB server run the below command.
```shell
docker-compose up
```
Run the backend API to serve the graph data with the below command.
```shell
cd ./graph-vis/backend && air
```
Run the frontend dev server to visualise the graph data with the below command.
```shell
cd ./graph-vis/frontend && npm run dev
```