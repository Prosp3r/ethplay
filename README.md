# ETHCALL AN JSONRPC API CALL SERVICE 

## Introduction
ETHCALL is a set of API end points designed to showcase examples of high load tollerant REST API service making queries to data loaded from back-end JSONRPC API calls. This is meant to be a **resilient, high throughput** API.

This means it can handle a fairly substantial number of API requests in very short time and carry out all that's required of it while maintaining a fairly stable uptime on minimal compute resources e.g. A single Linux server(wih fairly minimal memory resource), docker container, etc.

Its goal is to showcase how a simple design implemented in Go programming language can handle complex tasks in production easily.

After going through and carrying out the instructions in this document, you would have achieved the following.
+ Gained some understanding of how key aspects of the ETHCALL service code works.
+ Setup a simple Linux server or docker container to run the ETHCALL service.
+ Installed Go.(optional)
+ Retrived block data from the Ethereum mainnet through the Ethereum's Json RPC API
+ Ran a simulated load test using [ApacheBench](https://httpd.apache.org/)


## How it works

ETHCALL connects to the ethereum main net to retrieve the latest block every five seconds and will keep checking for new bocks until it reaches the limit set currently of 11 blocks.

To keep the data fresh i.e. as upto date as possible with the mainnet, persistence of block and transaction data has been avoided. So each time you fire up ETHCALL it starts a fresh set of JSON-RPC calls to the mainnet to get the next eleven newest blocks.

Some key features

 - ETHCALL checks and prevents duplication of blocks.
 - It keeps these in memory and maintains as often as "machinely" possible, a single read/write access to the storage.
 - It uses mutexes to protect memory reads and writes.
 



This API has four end points. I've described what each endpoint does in the table below.

<table> 
        <tr> <td> End Point </td><td> Function </td></tr>
        <tr> <td> / </td><td> Home endpoint - Displays a description of the other end-points </td></tr>
        <tr><td> /blocks </td><td> Blocks - to view a list of blocks fetched from the mainnet. Updates ever five seconds </td> </tr>
        <tr> <td> /transactions </td><td> Transactions - to view the JSON formated list of transactions carried out on the fetched block numbers </td></tr>
        <tr> <td> /receipts </td><td> Previous -  to view a JSON formated list of receipts </td></tr> 
</table>


#### System flow

![Ehcall system flow diagram](https://github.com/INFURA/INFRA-TEST-PROSPER-ONOGBERIE/blob/main/testshots/ethcallflow.png)



In the illustration above, points 1, 2, 3 and 4 are executed at the start of the system, point 5 is triggered by requests made on the API endpoints with read capabilities e.g. `/` , `/blocks` , `/transactions` and `/receipts`.

1. When the service is started, it opens a connection to the mainnet.
2. It then starts a Go routine that keeps updating the `blockStore` a simple map as in-memory store with the latest block number and hash from the mainnet(eth_getBlockByNumber). This pull continues until it gets the 11th block. After each block pull a cascade of transactions(eth_getTransactionByBlockNumberAndIndex) and transaction receipt (eth_getTransactionReceipt) pulls happen with the retrieved blockNumber and blockHash. 

3. The service also starts a http server using the gorilla mux package that listens for request on all endpoints.
4. When requests come in through any of the endpoints, the memory stores are queried, marshalled into JSON data and returned.

To prevent read conflicts in the in-memory store, all reads and writes are protected with the sync package mutex lock function.

NOTE: The API checks and error return code aspect of the above diagram were left there for formality as implementing them were not required in this case.
Return codes 200 and 500 howerver were implemented.



## Setting up ETHCALL

#### Prerequisites
+ An MacBook or Ubuntu Linux server or VPS with minimum 1024 MiB and 1 Core CPU.
+ Command line prompt access to the Ubuntu server
+ Docker installed and running on the MacBook or Server.

  ###### _A VPS could be acquired from any of the popular vendors like [Amazon AWS](https://aws.amazon.com), [DigitalOcean](https://digitalocean.com)_


_While these instructions may work on other Linux server types, I specify Ubuntu because the process has been thoroughly tested on Ubuntu(20.04) Linux servers._


#### Installation

To setup as user Ubuntu
Log on to your server through terminal with the user Ubuntu.


       $ cd ~
       $ docker run -d --name ethcall -p 8080:8080 sirpros/ethcall:latest

       or 
       On a mac with Go insalled.
       $ git clone <project-git-url>
       $ cd project-directory
       $ 

You can confirm that you have installed Git correctly by running the following url on an open browser:

        $ http://localhost:8080
. 

## Load testing with Apache Bench
To test the ability of our service to handle large amount of connections in fairly short time, we will be using [Apache Bench](https://httpd.apache.org/docs/2.4/programs/ab.html) utility which comes installed by default in some operating systems. 
We will need to install it on our server however.

Run the following command to install Apache Bench: 

        $ sudo apt install apache2-utils
        $ ab -c 10 -n 10000 -r http://localhost:8080/

The second command above sends ten thousand requests at 10 concurrent requests to the / endpoint.
You can vary the requests by changing the parameters.

#### Our results

When tested the service on a variety of systems, here are some results we got.

<table> 
        <tr><td> System configuration </td><td> Our Results </td><td>  </td>  </tr>
        <tr><td> MacBook Pro 8 GB 20.04[LTS](HVM) (0.5GiB, 1 vCPUs) </td><td> 4511.26 [#/sec], Avg time per request 0.241[ms] </td><td> </td>  </tr>
        <tr><td> Docker Container </td><td> 182.28 [#/sec], Avg time per request 5.486[ms] </td><td> </td>  </tr>
           
</table>


## Extras
  - Load performance on a single Docker container vary depending on the Docker host system capacity, requests throughput of between 560 to 9811 requests/second are achieved well for both non-write/read and write/read endpoints.
  The rates varies wide depending on the amount of CPU power being on host. 
