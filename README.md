# FairOS-dfs

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
##### user related APIs
- POST -F 'user=\<username\>' -F 'password=\<password\>' -F 'mnemonic=<12 words from bip39 list>' http://localhost:9090/v0/user/signup
- POST -F 'user=\<username\>' -F 'password=\<password\>' http://localhost:9090/v0/user/signup
- POST -F 'user=\<username\>' -F 'password=\<password\>' http://localhost:9090/v0/user/login 
- POST http://localhost:9090/v0/user/logout
- POST http://localhost:9090/v0/user/avatar
- POST -F 'first_name=\<firstName\>' -F 'middle_name=\<middleName\>' -F 'last_name=\<lastName\>' -F 'surname=\<surName\>' http://localhost:9090/v0/user/name
- POST -F 'phone=\<phone\>' -F 'mobile=\<mobile\>' -F 'address_line_1=\<address1\>' -F 'address_line_2=\<address2\>' http://localhost:9090/v0/user/contact
- POST http://localhost:9090/v0/user/export
- POST -F 'user=\<username\>' -F 'address=\<user_address\>'  -F 'password=\<password\>' http://localhost:9090/v0/user/import
- POST -F 'user=\<username\>' -F 'mnemonic=\<12_word_mnemonic\>'  -F 'password=\<password\>' http://localhost:9090/v0/user/import
- DELETE -F 'password=\<password\>' http://localhost:9090/v0/user/delete
- GET  -F 'user=\<username\>' http://localhost:9090/v0/user/present
- GET  -F 'user=\<username\>' http://localhost:9090/v0/user/isloggedin
- GET  http://localhost:9090/v0/user/stat
- GET  http://localhost:9090/v0/user/avatar
- GET  http://localhost:9090/v0/user/name
- GET  http://localhost:9090/v0/user/contact
- GET  http://localhost:9090/v0/user/share/inbox
- GET  http://localhost:9090/v0/user/share/outbox

##### pod related APIs   
- POST -F 'password=\<password\>' -F 'pod=\<podname\>'  http://localhost:9090/v0/pod/new
- POST -F 'password=\<password\>' -F 'pod=\<podname\>'  http://localhost:9090/v0/pod/open
- POST http://localhost:9090/v0/pod/sync
- POST http://localhost:9090/v0/pod/close
- DELETE http://localhost:9090/v0/pod/delete
- GET http://localhost:9090/v0/pod/ls
- GET -F 'user=\<username\>' -F 'pod=\<podname\>'  http://localhost:9090/v0/pod/stat

##### dir related APIs   
- POST -F 'dir=\<dir_with_path\>'  http://localhost:9090/v0/dir/mkdir
- DELETE -F 'dir=\<dir_with_path\>'  http://localhost:9090/v0/dir/rmdir
- GET  -F 'dir=\<dir_with_path\>'  http://localhost:9090/v0/dir/ls
- GET  -F 'dir=\<dir_with_path\>'  http://localhost:9090/v0/dir/stat

##### file related APIs   
- POST -F -H "fairOS-dfs-Compression: snappy/gzip" 'pod_dir=\<dir_with_path\>' -F 'block_size=\<in_Mb\>' -F 'files=@\<filename1\>' -F 'files=@\<filename2\>' http://localhost:9090/v0/file/upload  (compression header optional)
- POST -F 'file=\<file_path\>'  http://localhost:9090/v0/file/download
- POST -F 'file=\<file_path\>' -F 'to=\<destination_user_address\>' http://localhost:9090/v0/file/share
- POST -F 'ref=\<sharing_reference\>' -F 'dir=\<pod_dir_to_store_file\>' http://localhost:9090/v0/file/share/receive
- POST -F 'ref=\<sharing_reference\>' http://localhost:9090/v0/file/share/receiveinfo
- GET  -F 'file=\<file_path\>'  http://localhost:9090/v0/file/stat
- DELETE -F 'file=\<file_path\>'  http://localhost:9090/v0/file/delete

##### Key Value store related APIs
- POST -F 'file=\<kv table name\>' http://localhost:9090/v0/kv/new
- POST -F 'file=\<kv table name\>' http://localhost:9090/v0/kv/open
- POST -F 'file=\<kv table name\>' http://localhost:9090/v0/kv/count
- POST http://localhost:9090/v0/kv/ls
- DELETE -F 'file=\<kv table name\>' http://localhost:9090/v0/kv/delete
- POST -F 'file=\<kv table name\>' -F 'key=\<key\>' -F 'value=\<bytes\>' http://localhost:9090/v0/kv/entry/put
- GET -F 'file=\<kv table name\>' -F 'key=\<key\>' http://localhost:9090/v0/kv/entry/get
- DELETE -F 'file=\<kv table name\>' -F 'key=\<key\>' http://localhost:9090/v0/kv/entry/del
- POST -F 'file=\<kv table name\>' -F 'csv=@\<csv_file\>' http://localhost:9090/v0/kv/loadcsv
- POST -F 'file=\<kv table name\>' -F 'start=\<start_prefix\>' -F 'end=\<end\>' -F 'limit=\<no of records\>' http://localhost:9090/v0/kv/seek
- GET -F 'file=\<nkv table ame\>' http://localhost:9090/v0/kv/seek/getnext

##### Document store related APIs
- POST -F 'name=\<document table name\>' http://localhost:9090/v0/doc/new
- POST -F 'name=\<document table name\>' http://localhost:9090/v0/doc/open
- POST -F 'name=\<tdocument able name\>' -F 'expr=\<expression\>' http://localhost:9090/v0/doc/count
- POST -F 'name=\<document table name\>' -F 'expr=\<expression\>' -F 'limit=\<no of records\>' http://localhost:9090/v0/doc/find
- POST http://localhost:9090/v0/doc/ls
- DELETE -F 'name=\<tdocument able name\>' http://localhost:9090/v0/doc/delete
- POST -F 'name=\<tdocument able name\>' -F 'doc=\<json document in bytes\>' http://localhost:9090/v0/doc/entry/put
- GET -F 'name=\<document table name\>' -F 'id=\<document id\>' http://localhost:9090/v0/doc/entry/get
- DELETE -F 'name=\<document table name\>' -F 'id=\<document id\>' http://localhost:9090/v0/doc/entry/del
- POST -F 'name=\<document table name\>' -F 'json=@\<json_file\>' http://localhost:9090/v0/doc/loadjson

### REPL Commands in dfs
**dfs >>>** \<command\> where, \<command\> is listed below
##### user related commands
- user \<new\> (user-name) - create a new user and login as that user
- user \<del\> (user-name) - deletes a already created user
- user \<login\> (user-name) - login as a given user
- user \<logout\> (user-name) - logout as user
- user \<ls\> - lists all the user present in this instance
- user \<name\> (first_name) (middle_name) (last_name) (surname) - sets the user name information
- user \<name\> - gets the user name information
- user \<contact\> (phone) (mobile) (address_line1) (address_line2) (state) (zipcode) - sets the user contact information
- user \<contact\> gets the user contact information
- user \<share\> \<inbox\> - shows details of the files you have received from other users
- user \<share\> \<outbox\> - shows details of the files you have sent to other users
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
