
## dataflow

This project provides a data processing job and an HTTP interface.

Under the folder `cmd` there are two folders. Each includes main files of respective service.
`job` is a data processing pipeline and storage facility.  It retrieves a
number of JSONL files from AWS S3, parses,  and saves them in a Cassandra
database. Choosing a NoSQL database easily meets our needs, since the
only database query that we'll care for is just getting the row of a particular id.
We also want the database to perform well under massive write conditions.
`Cassandra` is a good choice here. `job` uses a cache. It is used to speed up the 
pipeline by eliminating the repetitive write operations to Cassandra, by
avoiding repeated calls to functions that process the same input.

`microservice` is a REST API that has only one endpoint:
For a request `/product/:id` returns Product as a `JSON` representation.  

## Project structure

1. There are application domain types under the main directory. `Fetch`,
   `Cache`, `ProductService`. 
2. Implementations of those types are in the packages. `cassandra`, `http`,
   etc.
3. Everything is wired together in the `cmd` subpackages. `cmd/job` &
   `cmd/microservice`.

### The application domains 

An application domain is a dependency that defines a type in an application,
independent of its implementation details. For example, a `Cache` describes what
it does without any implementation details.

We provide interfaces for managing the application domain data types, serving
as blueprints for underlying implementations. For instance, we define the
`ProductService` application domain interface for database operations, with
Cassandra serving as the actual implementation. 

This approach enables us to swap out implementations. An example is the `Fetch`
interface, which deals with getting resources from a filesystem. There are two
implementations of it in the project. One example (`S3FetchService`) fetches
from `S3`, and the other one (`FileReader`) reads files from the local filesystem. They have a
common interface, but one pulls data from the cloud, and the other reads a file from
the local filesystem. This makes testing and mocking services easier.

### Implementation

Some packages connect the application domain and the
concrete technology that we use to implement the domain. For instance
`cassandra.ProductService` implements `dataflow.ProductService` using
Cassandra.

The packages do not know each other, they communicate using
application domains. For example, the `http` package doesn't know which database it
should communicate with. It just uses an instance of `dataflow.ProductService`
application domain. 

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

These applications tie the services together depending on their case. 

`job` application creates a `s3` fetching service, then adds `cassandra`
storage layer and `redis` cache layer on top. Mechanically `job` pulls JSONL
files from AWS S3, processes them, and concurrently saves to the Cassandra. 
It exits if the pipeline succeeds or is interrupted by a signal or error.   

`microservice` application in comparison to the `job` creates a `http` service and
wires it up only with the `cassandra` storage layer. It is a daemon process that
listens a particular port for incoming HTTP requests..  


### Setup:
 There are 4 files in an AWS S3 bucket with names `products-[1..4].jsonl`. 
 They are present in region  `eu-central-1` under bucket: `casestudy`. 
 The task is: 
 > Fetch those files, process the JSONL files, and save
 them into storage database. Take optimization measures if necessary. Serve them in
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
Redis instance in the local environment using [Testcontainers for
Go](https://golang.testcontainers.org/).  
To run all integration tests (database
tests) and unit tests, call the following on the command line.
>  The project needs `docker` for integration tests in some packages. In some
>  GNU/Linux systems, you might need privileged access to the `docker`.

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
We need to take one more step to work with Cassandra. 
We call the following script under the `scripts` directory. 
This will create a keyspace and table on the database.

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
Optionally you can give `-concurrency n` flag to the CLI. 
This regulates the number of concurrent executions of some
stages of the pipeline. Default is the number of logical cores of the CPU of the host
machine. It is better to leave the default value here. It can be used to measure the relationship 
between concurrency and performance. 
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

### Optimization using cache 
To see and appreciate the achievement cashing via redis, we can do a demonstration. 
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

### Discussion

The pipeline stage can be more performant by introducing more concurrency.   

Currently, `Split` and `ConvertJSON` methods leverage just one goroutine 
in contrast to other stages of the pipeline,   
which are fully concurrent. Using a `fan-out`-`fan-in` concurrency pattern in the split 
and conversion stages might help optimize the execution time of the pipeline. 
My initial attempts and observations didn't show much improvement, so I abandoned development. 
The conclusion I reached was, that for relatively small number of files and lines, as in our current case, 
this optimization might not benefit too much. 



