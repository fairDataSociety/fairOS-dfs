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
word mnemonic based hd wallet. This wallet is passwod protected and stored in 
the datadir. whenever a user created a pod for himself, a new key pair is created 
using this mnemonic. A user can use this mnemonic and import their account in any 
device and instantly see all their pods.

### What is a pod?
A pod is a personal drive created by a user in fairOS-dfs. It is used to store files and related metadata in a decentralised fashion. A pod is always under the control of the user who created it. A user can create and store any number of files or directories in a pod. 
The user can share files in his pod with any other user just like in other centralised drives like dropbox. Not only users, a pod can be used by decentralised applications (DApp's) to store data related to that user.

Pod creation is cheap. A user can create multiple pods and use it to organise his data. for ex: Personal-Pod, Applications-Pod etc.

### How to run FairOS-dfs?
Download the latest release from https://github.com/fairDataSociety/fairOS-dfs/releases.

Or use Docker to run the project https://docs.fairos.fairdatasociety.org/docs/fairOS-dfs/docker-installation.

Or build the latest version with the instruction https://docs.fairos.fairdatasociety.org/docs/fairOS-dfs/manual-installation.

### Demo 1: FairOS-dfs Introduction
[![](https://j.gifs.com/lx3x0l.gif)](https://gateway.ethswarm.org/access/541f55413e02774c9446525d0cf3a92984cc541e4d9e73cb70c1dabe2e870bc5)
### Demo 2: FairOS-dfs File Sharing
[![](https://j.gifs.com/vl3l5g.gif)](https://gateway.ethswarm.org/access/834191ac103224cd2c665a34f2eb5113926e6624adbdddfc7a86f314eb7cfeeb)
### Demo 3: FairOS-dfs Adding a new Device
[![](https://j.gifs.com/D1g1rY.gif)](https://gateway.ethswarm.org/access/7a8964194ffb923b98cc60711ff1925d2411537fc9f2dc80ee9219a49d0e4949)
### Demo 4: Introdution to Key Value Store over Swarm
[![](https://j.gifs.com/6XZwvl.gif)](https://gateway.ethswarm.org/access/130dcf7d01442836bc14c8c38db32ebfc4d5771c28677438b6a2a2a078bd1414)
### Demo 5: Adding large datasets in KV store in Swarm
[![](https://j.gifs.com/jZDwkl.gif)](https://gateway.ethswarm.org/access/2688969c020cb736afae9b2f6d65c834414f83f8b4fdced077eb3e5f9a7266af)



### HTTP APIs

https://docs.fairos.fairdatasociety.org/docs/fairOS-dfs/api-reference


### REPL Commands in dfs-cli

https://docs.fairos.fairdatasociety.org/docs/fairOS-dfs/cli-reference

### Make binaries for all platforms

To make binaries for all platforms run this command

`./generate-exe.sh`
