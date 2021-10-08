# FairOS-dfs

Latest documentation is available  @ ![https://fairos.io](https://fairos.io)

![FairOS-dfs](https://github.com/fairDataSociety/fairOS-dfs/blob/master/docs/images/FairOS-dfs.png)

The Decentralised File System (dfs) is a file system built for the ![FairOS](https://github.com/fairDataSociety/fairOS/blob/master/README.md).
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

Pod cretion is cheap. A user can create multiple pods and use it to organise his data. for ex: Personal-Pod, Applications-Pod etc.

### How to build and run dfs?
- git clone https://github.com/fairdatasociety/fairOS-dfs.git
- cd fairOS-dfs
- make binary
- ./dist/dfs prompt (starts dfs in REPL mode)
- ./dist/dfs server (starts dfs in server mode serving API in port 9090 by default)

### Demo 1: FairOS-dfs Introduction
[![](https://j.gifs.com/lx3x0l.gif)](https://bee.fairos.io/files/19c1bd8c5714db4f798e07421dc7a20497537e1d1d8ad4f95cfcac8775effd05)
### Demo 2: FairOS-dfs File Sharing
[![](https://j.gifs.com/vl3l5g.gif)](https://bee.fairos.io/files/43a9a08b4ffa7aea1f1d0f0ba0e0a239e6e14bed5b5f4145821a81489d43443e)
### Demo 3: FairOS-dfs Adding a new Device
[![](https://j.gifs.com/D1g1rY.gif)](https://bee.fairos.io/files/5b32278f5d7dbd69f08090a375ab3597956ec329d0dc14a016d8553a1e871eb0)
### Demo 4: Introdution to Key Value Store over Swarm
[![](https://j.gifs.com/6XZwvl.gif)](https://bee.fairos.io/files/94a5d957a90b52be9eab73d61c7c3f5a68848d42c5266c25d7b056bb3871d9ce)
### Demo 5: Adding large datasets in KV store in Swarm
[![](https://j.gifs.com/jZDwkl.gif)](https://bee.fairos.io/files/e44f2914a19a789a7a0fff4fc37e3243dadbda5e4e8e1c9a860ef1edf8d4474e)



### HTTP APIs

https://fairos.io/bzz/fairos.eth/docs/fairOS-dfs/api-reference/


### REPL Commands in dfs-cli
**dfs-cli >>>** \<command\> where, \<command\> is listed below
##### user related commands
- user \<new\> (user-name) - create a new user and login as that user
- user \<del\> (user-name) - deletes a already created user
- user \<login\> (user-name) - login as a given user
- user \<logout\> (user-name) - logout as user
- user \<ls\> - lists all the user present in this instance
- user \<export\> - exports the given user
- user \<import\> (user-name) (address) - imports the user to another device
- user \<import\> (user-name) (12 word mnemonic) - imports the user if the device is lost"
- user \<stat\> - shows information about a user
##### pod related commands
- pod \<new\> (pod-name) - create a new pod and login to that pod
- pod \<del\> (pod-name) - deletes a already created pod
- pod \<login\> (pod-name) - login to a already created pod
- pod \<stat\> (pod-name) - display meta information about a pod
- pod \<sync\> (pod-name) - sync the contents of a logged in pod from Swarm
- pod \<logout\>  - logout of a logged in pod
- pod \<ls\> - lists all the pods created for this account
##### directory & file related commands
- cd \<directory name\>
- ls 
- copyToLocal \<source file in pod, destination directory in local fs\>
- copyFromLocal \<source file in local fs, destination directory in pod, block size in MB\>
- mkdir \<directory name\>
- rmdir \<directory name\>
- rm \<file name\>
= pwd - show present working directory
- head \<no of lines\>
- cat  - stream the file to stdout
- stat \<file name or directory name\> - shows the information about a file or directory
- share \<file name\> -  shares a file with another user
- receive \<sharing reference\> \<pod dir\> - receives a file from another user
- receiveinfo \<sharing reference\> - shows the received file info before accepting the receive 
##### Key Value store commands
- kv \<new\> (table-name) - create new key value store
- kv \<delete\> (table-name) - delete the  key value store
- kv \<ls\> - lists all the key value stores
- kv \<open\> (table-name) - open already created key value store
- kv \<get\> table-name) (key) - get value from key
- kv \<put\> (table-name) (key) (value) - put key and value in kv store"
- kv \<del\> (table-name) (key) - delete key and value from the store
- kv \<loadcsv\> (table-name) (local csv file) - loads the csv file in to kv store
- kv \<seek\> (table-name) (start-key) (end-key) (limit) - seek to the given start prefix
- kv \<getnext\> (table-name) - get the next element
##### Document store commands
- doc \<new\> (table-name) (si=indexes) - creates a new document store
- doc \<delete\> (table-name) - deletes a document store
- doc \<open\> (table-name) - open the document store
- doc \<ls\>  - list all document dbs
- doc \<count\> (table-name) (expr) - count the docs in the table satisfying the expression
- doc \<find\> (table-name) (expr) (limit)- find the docs in the table satisfying the expression and limit
- doc \<put\> (table-name) (json) - insert a json document in to document store
- doc \<get\> (table-name) (id) - get the document having the id from the store
- doc \<del\> (table-name) (id) - delete the document having the id from the store
- doc \<loadjson\> (table-name) (local json file) - load the json file in to the newly created document db  
##### management commands
- help - display this help
- exit - exits from the prompt

### Make binaries for all platforms

To make binaries for all platforms run this command

`./generate-exe.sh`
