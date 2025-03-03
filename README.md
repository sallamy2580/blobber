
[![Build](https://github.com/0chain/blobber/actions/workflows/build-&-publish-docker-image.yml/badge.svg)](https://github.com/0chain/blobber/actions/workflows/build-&-publish-docker-image.yml)
[![Test](https://github.com/0chain/blobber/actions/workflows/tests.yml/badge.svg)](https://github.com/0chain/blobber/actions/workflows/tests.yml)
[![GoDoc](https://godoc.org/github.com/0chain/blobber?status.png)](https://godoc.org/github.com/0chain/blobber)
[![codecov](https://codecov.io/gh/0chain/blobber/branch/staging/graph/badge.svg)](https://codecov.io/gh/0chain/blobber)

# Blobber Setup
This readme provide instructions on how to run blobber locally

## Table of Contents  

- [Initial Setup](#initial-setup) - [Directory Setup for Blobbers](#directory-setup-for-blobbers)

- [Building and Starting the Blobber](#building-and-starting-the-nodes) 

- [Troubleshooting](#troubleshooting)

- [Connect to other network](#connect-to-other-network)

- [Miscellaneous](#miscellaneous) - [Cleanup](#cleanup) - [Minio Setup](#minio)

- [Run blobber on ec2 / vm / bare metal](https://github.com/0chain/blobber/blob/master/docker.aws/README.md)

- [Run blobber on ec2 / vm / bare metal over https](https://github.com/0chain/blobber/blob/master/https/README.md)

- [Blobber local development guideline](dev.local/README.md)

## Initial Setup

### Required OS and Software Dependencies

- Linux (Tested on Ubuntu Desktop 20.02)
- MacOS
- Docker([Link](https://docs.docker.com/engine/install/)
- Docker Compose([Link](https://docs.docker.com/compose/install/))  

### Directory Setup for Blobbers

1. Clone the Blobber repository using the command
```
git clone https://github.com/0chain/blobber.git
```
2. In the git/blobber run the following command
  
```
./docker.local/bin/blobber.init.setup.sh
```

## Building and Starting the Nodes
  
1. Setup a network called testnet0 for each of these node containers to talk to each other.
 
 ```
docker network create --driver=bridge --subnet=198.18.0.0/15 --gateway=198.18.0.255 testnet0
```
Note: Run all scripts as sudo  

2. Set up the block_worker URL

A block worker URL is a field in the `blobber/config/0chain_validator.yaml` and `blobber/config/0chain_blobber.yaml` configuration files that require the URL of bloockchain network you want to connect to.For testing purposes we will connect to the beta 0chain network and replace the default URL in blobber/config/0chain_validator.yaml and 0chain_blobber.yaml with the below-mentioned URL.
```
block_worker: http://beta.0chain.net/dns
```

3. Go back to the blobber directory and build blobber containers using the scripts below
```
./docker.local/bin/build.base.sh
./docker.local/bin/build.blobber.sh
./docker.local/bin/build.validator.sh
```
Note: Run all scripts as sudo. 
This would take few minutes.

To link to local gosdk so that the changes are reflected on the blobber build please use the below command(optional)

```
./docker.local/bin/build.blobber.dev.sh

```
For Mac with Apple M1 chip use the following [guide](https://github.com/0chain/blobber/blob/staging/dev.local/README.md) to build and start blobbers.

Now register a Wallet using zboxcli to perform storage operations on blobbers.Build instructions for zbox are [here](https://github.com/0chain/zboxcli#installation-guides)

5. Verify whether Zbox has properly build by running the following command
```
./zbox
```

6. To register a wallet on Zbox to be used both by the blockchain and blobbers. Use the following Zbox command
 ```
./zbox register
```
Successful Response:
```
Wallet Registered
```
7. Now navigate to the .zcn folder (this is created during zbox build) 
```
cd $HOME/.zcn/
```
8. Open the wallet.json file. It should be similar to the similar to the output below:
```
{"client_id":"4af719e1fdb6244159f17922382f162387bae3708250cab6bc1c20cd85fb594c",
"client_key":"da1769bd0203b9c84dc19846ed94155b58d1ffeb3bbe35d38db5bf2fddf5a91c91b22bc7c89dd87e1f1fecbb17ef0db93517dd3886a64274997ea46824d2c119","keys":[{"public_key":"da1769bd0203b9c84dc19846ed94155b58d1ffeb3bbe35d38db5bf2fddf5a91c91b22bc7c89dd87e1f1fecbb17ef0db93517dd3886a64274997ea46824d2c1>
"private_key":"542f6be49108f52203ce75222601397aad32e554451371581ba0eca56b093d19"}],"mnemonics":"butter whisper wheat hope duck mention bird half wedding aim good regret maximum illegal much inch immune unlock resource congress drift>
"version":"1.0","date_created":"2021-09-09T20:22:56+05:30"}
```
9. Copy the client_id value and paste it into blobbers and validators settings. The files can be found in `blobber/config` directory.
  
10. Open both the `blobber/config/0chain_validator.yaml` and `blobber/config/0chain_blobber.yaml` and edit the `delegate_wallet` value with your `client_id` value.

11 Now run the blobbers by navigating into blobber directories for Blobber1 (git/blobber/docker.local/blobber1) and run the container using

```
# For locally build images
../bin/blobber.start_bls.sh

# For remote images
../bin/p0blobber.start.sh

```
**_Note: Replace the localhost form `docker.local/p0docker-compose.yml` to your public IP if you are trying to connect to another network ._**

## Troubleshooting

12. Ensure the port mapping is all correct:

```
docker ps

```
This should display the container image blobber_blobber and should have the ports mapped like "0.0.0.0:5050->5050/tcp"

13. Now check whether the blobber has registered to the blockchain by running the following zbox command

```
./zbox ls-blobbers
```
In the response you should see the local blobbers mentioned with their urls for example http://198.18.0.91:5051 and http://198.18.0.92:5052

Sample Response:
```
- id:                    0bf5ae461d6474ca1bebba028ea57d646043bbfb6a4188348fd649f0deec5df2
  url:                   http://beta.0chain.net:31304
  used / total capacity: 14.0 GiB / 100.0 GiB
  last_health_check:	  1635347306
  terms:
    read_price:          26.874 mZCN / GB
    write_price:         26.874 mZCN / GB / time_unit
    min_lock_demand:     0.1
    cct:                 2m0s
    max_offer_duration:  744h0m0s
- id:                    7a90e6790bcd3d78422d7a230390edc102870fe58c15472073922024985b1c7d
  url:                   http://198.18.0.92:5052
  used / total capacity: 0 B / 1.0 GiB
  last_health_check:	  1635347427
  terms:
    read_price:          10.000 mZCN / GB
    write_price:         100.000 mZCN / GB / time_unit
    min_lock_demand:     0.1
    cct:                 2m0s
    max_offer_duration:  744h0m0s
- id:                    f65af5d64000c7cd2883f4910eb69086f9d6e6635c744e62afcfab58b938ee25
  url:                   http://198.18.0.91:5051
  used / total capacity: 0 B / 1.0 GiB
  last_health_check:	  1635347950
  terms:
    read_price:          10.000 mZCN / GB
    write_price:         100.000 mZCN / GB / time_unit
    min_lock_demand:     0.1
    cct:                 2m0s
    max_offer_duration:  744h0m0s
- id:                    f8dc4aaf3bb32ae0f4ed575dd6931a42b75e546e07cb37a6e1c6aaf1225891c5
  url:                   http://beta.0chain.net:31305
  used / total capacity: 13.3 GiB / 100.0 GiB
  last_health_check:	  1635347346
  terms:
    read_price:          26.874 mZCN / GB
    write_price:         26.865 mZCN / GB / time_unit
    min_lock_demand:     0.1
    cct:                 2m0s
    max_offer_duration:  744h0m0s
```

Note: When starting multiple blobbers, it could happen that blobbers are not being registered properly (not returned on `zbox ls-blobbers`). 
   
Blobber registration takes sometime and adding at least 5 second wait before starting the next blobber usually avoids the issue.
  
14. Now you can create allocations on blobber and store files. 

Note: If unable to create new allocations as shown below.

```
./zbox newallocation --lock 0.5
Error creating allocation: transaction_not_found: Transaction was not found on any of the sharders
```

To fix this issue you must lock some tokens on the blobber.Get the local blobber id using the `./zbox ls-blobbers` and use the following command 

```
zbox sp-lock --blobber_id f65af5d64000c7cd2883f4910eb69086f9d6e6635c744e62afcfab58b938ee25 --tokens 0.5
```

## Connect to other network


- Your network connection depends on the block_worker url you give in the `config/0chain_blobber/validator.yaml` and `0chain_blobber.yaml` config file.
 
```

block_worker: http://198.18.0.98:9091

```

This works as a dns service, You need to know the above url for any network you want to connect, Just replace it in the above mentioned file.

For example: If you want to connect to test network
  
```

block_worker: https://test.0chain.net/dns

```


## Miscellaneous
 
### Cleanup


1. Get rid of old unused docker resources:

  

```

docker system prune

```

  

2. To get rid of all the docker resources and start afresh:

  

```

docker system prune -a

```

  

3. Stop All Containers

  

```

docker stop $(docker ps -a -q)

```

  

4. Remove All Containers

  

```

docker rm $(docker ps -a -q)

```

  

### Minio

  

- You can use the inbuild minio support to store old data on cloud

  

You have to update minio_config file with the cloud creds data, The file can found at `docker.local/keys_config/minio_config.txt`.

The following order is used for the content :

  

```

CONNECTION_URL

ACCESS_KEY_ID

SECRET_ACCESS_KEY

BUCKET_NAME

REGION

```

  

- Your minio config file is then used in the docker-compose while starting the sharder node

  

```

--minio_file keysconfig/minio_config.txt

```

  

- You can either update the setting in the same file which is given above or create a new one with you config and use that as

  

```

--minio_file keysconfig/your_new_minio_config_file.txt

```

  

\*\*\_Note: Do not forget to put the file in the same config folder OR mount your new folder.

  

- Apart from private connection config, There are other options as well in the 0chain_blobber.yaml file to manage minio settings.

  

Sample config

  

```

minio:

# Enable or disable minio backup service

start: false

# The frequency at which the worker should look for files, Ex: 3600 means it will run every 3600 seconds

worker_frequency: 3600 # In Seconds

# Use SSL for connection or not

use_ssl: false

```

  

- You can also tweak the cold storage setting depending how you want to decide which data to move to the cloud.

  

Sample config

  

```

cold_storage:

# Minimum file size to be considered for moving to cloud

min_file_size: 1048576 #in bytes

# Minimum time for which file is not updated or not used

file_time_limit_in_hours: 720 #in hours

# Number of files to be queried and processed at once

job_query_limit: 100

# Capacity in percentage after which the cloud backup should start work

max_capacity_percentage: 50

# Delete local copy once the file is moved to cloud

delete_local_copy: true

# Delete cloud copy if the file is deleted from the blobber by user/other process

delete_cloud_copy: true

```