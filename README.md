# FairOS-dfs

Latest documentation is available at [https://docs.fairos.fairdatasociety.org/docs/](https://docs.fairos.fairdatasociety.org/docs/)

![FairOS-dfs](https://github.com/fairDataSociety/fairOS-dfs/blob/master/docs/images/FairOS-dfs.png)

The Decentralised File System (dfs) is a file system built for the [FairOS](https://github.com/fairDataSociety/fairOS/blob/master/README.md).
It is a stateless thin layer which uses the building blocks provided by Swarm to provide high level functionalities like
- Exposing a logical file system
- Creation of logical drives
- User and Permission management
- Charging and Payments
- Mutable, Indexed data structures over immmutable file system

dfs can be used for the following use cases
1) Personal data store
2) Application data store (for both Web 3.0 DApps and web 2.0 Apps)
3) Data sharing with single user and on an organizational level

### User
The first step in dfs is to create a user. Every user is associated with a 12 
word mnemonic based hd wallet. This wallet is password protected and stored in 
the datadir. whenever a user created a pod for himself, a new key pair is created 
using this mnemonic. A user can use this mnemonic and import their account in any 
device and instantly see all their pods.

### What is a pod?
A pod is a personal drive created by a user in fairOS-dfs. It is used to store files and related metadata in a decentralised fashion. A pod is always under the control of the user who created it. A user can create and store any number of files or directories in a pod. 
The user can share files in his pod with any other user just like in other centralised drives like dropbox. Not only users, a pod can be used by decentralised applications (DApp's) to store data related to that user.

Pod creation is cheap. A user can create multiple pods and use it to organise his data. for ex: Personal-Pod, Applications-Pod etc.

### How to run FairOS-dfs?
Run the following command to download the latest release

```
curl -o- https://raw.githubusercontent.com/fairDataSociety/fairOS-dfs/development.1/download.sh | bash
```
```
wget -qO- https://raw.githubusercontent.com/fairDataSociety/fairOS-dfs/development.1/download.sh | bash
```

Or download the latest release from https://github.com/fairDataSociety/fairOS-dfs/releases.

Or use Docker to run the project https://docs.fairos.fairdatasociety.org/docs/fairOS-dfs/docker-installation.

Or build the latest version with the instruction https://docs.fairos.fairdatasociety.org/docs/fairOS-dfs/manual-installation.

### Configure FairOS-dfs
To get the most out of your FairOS-dfs it is important that you configure FairOS-dfs for your specific use case!

##### Configuration for Bee 
```
bee:
  bee-api-endpoint: http://localhost:1633
  bee-debug-api-endpoint: http://localhost:1635
  postage-batch-id: ""
```

##### Configuration for FairOS-dfs
```
dfs:
  data-dir: /Users/fairos/.fairOS/dfs
  ports:
    http-port: :9090
    pprof-port: :9091
```

#### Other configuration
```
cookie-domain: api.fairos.io
cors-allowed-origins: []
verbosity: trace
```

Run `dfs config` to see all configurations

### Introduction to Key Value Store over Swarm
[![](https://j.gifs.com/6XZwvl.gif)](https://gateway.ethswarm.org/access/130dcf7d01442836bc14c8c38db32ebfc4d5771c28677438b6a2a2a078bd1414)

### HTTP APIs

https://docs.fairos.fairdatasociety.org/docs/fairOS-dfs/api-reference

### REPL Commands in dfs-cli

https://docs.fairos.fairdatasociety.org/docs/fairOS-dfs/cli-reference

### Make binaries for all platforms

To make binaries for all platforms run this command

`./generate-exe.sh`

