# FairOS-dfs
The Decentralised File System (dfs) is a file system built for the FairOS.
It is a stateless thin layer which uses the building blocks provided by Swarm to 
provide high level functionalities like
- Exposing a logical file system
- Creation of logical drives
- User and Permission management
- Charging and Payments
- Mutable, Indexed data structures over immmutable file system

dfs has the fllowing usecases
1) Personal data store
2) Application data store (for both Web 3.0 DApps and web 2.0 Apps)
3) Data sharing with single user and on an organizational level


![FairOS-dfs](https://github.com/fairDataSociety/fairOS-dfs/blob/master/docs/images/FairOS-dfs.png)

### User
The first step in dfs is to create a user. Every user is associated with a 12 
word mnemonic based hd wallet. This wallet is passwod protected and stored in 
the datadir. whenever a user created a pod for himself, a new key pair is created 
using this mnemonic. A user can use this mnemonic and import their account in any 
device and instantly see all their pods.

### What is a pod?
A pod is a personal drive created by a user in fairOS-dfs. It is used to store files and related metadata in a decentralised fashion. A pod is always under the control of the user who created it. A user can create store any number of files or directories in a pod. 
The user can share files in his pod with any other user just like in other centralised drives like dropbox. Not only users, a pod can be used by decentralised applications (DApp's) to store data related to that user.

The basic storage unit in dfs is a pod. A user can create multiple pods and use it to organise their data. for ex: Personal-Pod, Applications-Pod etc.

### How to build and run dfs?
- git clone https://github.com/fairdatasociety/fairOS-dfs.git
- cd fairOS-dfs
- make binary
- ./dist/dfs prompt (starts dfs in REPL mode)
- ./dist/dfs server (starts dfs in server mode serving API in port 9090 by default)

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
##### management commands
- help - display this help
- exit - exits from the prompt
