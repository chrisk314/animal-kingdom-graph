# Animal Kingdom Graph
This project contains tools to visualise the animal kingdom as a graph. There are two components: a Wikipedia scraper; and a graph visualisation web app.

## Scraper
A Wikipedia scraper implemented in Golang which builds a graph of 
the animal kingdom. The Colly web crawler package and Goquery package are used to 
traverse Wikipedia pages related to animal species. Data from each page related
to an animal species is stored in an Arango graph database along with the related 
taxonomic hierarchy.

There are many animal species which do not fit neatly into a consistent taxonomic heirarchy. 
To simplify things, only species whose taxonomic heirarchy matches the below sequence are 
extracted and stored.

kingdom -> phylum -> class -> order -> family -> genus -> species

### Install
To install the dependencies and build the binary run the below commands from the [wiki-scraper](./wiki-scraper/) directory.
```shell
go get
go build
```

### Run
To run an Arango DB server run the below command.
```shell
docker-compose up
```
Now the scraper can be launched with the below command.
```shell
cd ./wiki-scraper && wiki_scraper
```

## Visualiser
A graph visualiser implemented as a backend Golang API to serve data from the ArangoDB database,
and a Vue.js SPA to visualise the graph data interactively with Cytoscape.js. Clicking on nodes
in the graph will retrieve child nodes from the backend API or cache and display them. Clicking
an already populated node will remove the child nodes from the graph.

### Install
Install the backend API dependencies.
```shell
pushd graph-vis/backend
go get
popd
```
Install the frontend API dependencies.
```shell
pushd graph-vis/frontend
npm install
popd
```

### Run
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