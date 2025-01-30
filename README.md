
## dataflow

This project provides a data processing job and an HTTP API to serve
data included in databases.

Under the folder `cmd` there are two folders. Each includes main files of respective service.
`job` is a data processing pipeline and storage facility.  It retrieves a
number of JSONL files from AWS S3, parses them,  and saves the parsed data in a Cassandra
table.

`microservice` is a REST API that has only one endpoint:
For a request `/product/:id` returns Product as a `JSON` representation.  

## Project structure

1. Abstract Interfaces are defined in the files under the main directory.  
2. Implementations of those types are in the packages. `cassandra`, `http`,
   etc.
3. Everything is wired together in the `cmd` subpackages. `cmd/job` &
   `cmd/microservice`.

### The application domains 

We provide interfaces for managing the application domain data types, serving
as blueprints for underlying implementations. For instance, we define the
`ProductService` application domain interface for the database operations on product type,
  whereas `cassandra` serving as the actual implementation. 
 This makes testing and mocking services easier.

### Implementation


The packages do not know each other, they communicate using
application domains. the `http` package doesn't know which database it
should communicate with. It just uses an instance of `dataflow.ProductService` interface.

- `http`: Implements product service over HTTP.
- `cassandra`: Implements product service storage layer. 
- `redis`: Implements product service cache layer. 
- `mock`: simple mock to enable `http` unit tests in isolation 
- `s3`: Implements fetch service for `S3`.


## Case Study

### Project binaries

The subpackages are wired together under `cmd/job` and `cmd/microservice` to
make working applications. 

- `job`: The pipeline CLI. 
- `microservice`: The HTTP server.

### Setup:
 Let's assume, there are 4 files in an AWS S3 bucket with names `products-[1..4].jsonl`. 
 They are present in region  `eu-central-1` under bucket: `casestudy`. 
 The task is: 
 > Fetch those files, process the JSONL files, and save the result into storage database.
  Take optimization measures if necessary. Serve them in
 terms of an HTTP based microservice. For a given product ID, the service should
 return the details of the relevant product . 

#### Configure AWS Configuration 

To be able to use AWS Go SDK we need to provide environment variables. 
```sh
export AWS_S3_REGION="eu-central-1" 
export AWS_S3_BUCKET="casestudy"
export AWS_ACCESS_KEY_ID=YOUR_AKID 
export AWS_SECRET_ACCESS_KEY=YOUR_SECRET_KEY
```

#### Tests 

This command pulls docker images, and deploys a Cassandra instance and a
Redis instance in the local environment using [Testcontainers for Go](https://golang.testcontainers.org/).  
To run all integration tests (database tests) and unit tests, call the following on the command line.

```sh 
go test ./... -race
```

### Usage 
We need `docker` and `docker-compose` for running `job` and `microservice` command line applications 
in terminal environments.

Both applications run on the command line and they both need a configuration file.

One configuration file is supplied as an example in the project root root, `dataflow.conf`. 

> Warning: Configuration files might contain sensitive credentials. 
  Additional steps might have to be taken for production systems.

The project requires a running Cassandra database instance and a Redis database. 
Call this under the project folder setup them on your local environment. 

```sh 
docker-compose up -d 
```
Cassandra needs a bit of time to establish its configuration. 
The following command will help to find out if it is ready. 
```sh 
docker exec -it cassandra-service  cqlsh -e "describe keyspaces"
``` 
Our database is ready, if there is no error. 
The following line will create a keyspace and table on the database.

```sh 
sudo sh ./scripts/provision.sh 
```

Now the setup is ready for our projects. Let's run the `job` command first. This
command pulls all the product files from AWS, processes them via a pipeline, and
stores them in a database. The first run might take a bit longer since we don't 
profit from the cache at the first run.

```sh 
go run cmd/job/main.go -config dataflow.conf 
```

For example, if we try the following command, we'll get a slower duration of execution. 
```sh 
go run cmd/job/main.go -config dataflow.conf -concurrency 1
```

Start the `microservice` HTTP daemon. 
```sh 

go run cmd/microservice/main.go -config dataflow.conf 
```

Test it: 
```sh 

curl localhost:8080/product/42
```

### Using `redis` cache 
To look into caching via redis, we can do a demonstration. 
First on the project directory call the following
command to remove the database instances. 
```sh 
docker-compose down --remove-orphans

``` 
Then re-create fresh instances of Cassandra and Redis..
```
sh docker-compose up -d 
``` 
Wait  like 30 seconds to let Cassandra stand up. 
Use this as a readiness probe: 
```sh 

docker exec -it cassandra-service  cqlsh -e "describe keyspaces" 
```

After this, we have a pipeline setup with an empty cache layer.  
This means the data processing should take a bit longer. Let's check.   
This command measures how much time elapsed for the execution of the job. 

```sh 
go run cmd/job/main.go -config dataflow.conf 
``` 
I got the following result on my computer.

```sh 
Pipeline tooks 29.788467364s 
```

After the second execution of the same command, I got the following output on my local.
```sh 

Pipeline tooks 15.399257129s 
```
This result shows the efficiency of the optimization through caching. 
We save half of the execution time of the pipeline.





