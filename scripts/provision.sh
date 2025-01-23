#!/bin/bash
set -e

docker exec -it cassandra-service  cqlsh -e "describe keyspaces"

docker exec -it cassandra-service  cqlsh -e  "create keyspace case_study_devel with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 }";

#create products table on the casestudy database
docker exec -it cassandra-service  cqlsh -e  "create table case_study_devel.products(id int,
                                title text, 
                                price float, 
                                category text,
                                 brand text, 
                                 url text, 
                                 description text,  
                                 PRIMARY KEY(id));"

#If you're done with testing, you can release the resources with:
# docker-compose down --remove-orphans
