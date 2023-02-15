# fairOS-dfs-wasm

fairOS-dfs now Supports wasm

## Build
```
make build-all
```

This will create two binaries in `dist` folder
20665380 fairos.wasm
5721514  fairos.wasm.gz // a gzip compressed wasm binary

## Example usage

### Running wasm
```
<html>
   	<head>
   	    /*
   	     * Download https://github.com/fairDataSociety/fairOS-dfs-wasm/blob/main/example/wasm_exec.js in your project
   	     * this is also availabe in your go root. `$(go env GOROOT)/misc/wasm/wasm_exec.js`
   	     * load wasm_exec.js before loading wasm itself, this is required for wasm build with golang
   	     */
		<script src="wasm_exec.js"></script>
		
		// Load wasm
		<script>
			const go = new Go();
	  
			let mod, inst, sessionID;
	  
			WebAssembly.instantiateStreaming(fetch("fairos.wasm"), go.importObject).then(
			    result => {
                    mod = result.module;
                    inst = result.instance;
                    go.run(inst).then( r => {
                        console.log("exiting...")
                    })
			    }
			);
		  </script>
   	</head>
   
	...
</html>
```

### Running gzipped wasm
```
<html>
   	<head>
   	    /*
   	     * Download https://github.com/fairDataSociety/fairOS-dfs-wasm/blob/main/example/wasm_exec.js in your project
   	     * this is also availabe in your go root. `$(go env GOROOT)/misc/wasm/wasm_exec.js`
   	     * load wasm_exec.js before loading wasm itself, this is required for wasm build with golang
   	     */ 
		<script src="wasm_exec.js"></script>
		
		// We can use pako to load compressed wasm
		<script src="pako.min.js"></script>
		
		// Load wasm
		<script>
			const go = new Go();
	  
			let mod, inst, sessionID;
	  
			const go = new Go();
			let sessionID;
			fetch("fairos.wasm.gz").then( r => {
				r.arrayBuffer().then( async b => {
					let buffer = pako.ungzip(b);

					if (buffer[0] === 0x1f && buffer[1] === 0x8b) {
						buffer = pako.ungzip(buffer);
					}
					const result = await WebAssembly.instantiate(buffer, go.importObject);
					go.run(result.instance).then( r => {
						console.log("exiting...")
					})
				})
			})
		  </script>
   	</head>
   
	...
</html>
```

## Api

### connect - Create fairOS api instance. This will verify if we can connect with the given bee
```
let resp = await connect("BEE_API", "BATCH_ID", "IS_USING_BEE_GATEWAY_PROXY", "RPC_ENDPOINT", "NETWORK")
console.log(resp)
```

### stop - exit the program
```
stop() 
```

### login
```
let resp = await login(USER, PASSWORD)
let r = JSON.parse(resp)
console.log(r.sessionId)
```

Note: fairOS-dfs maintains state for a logged-in user. we have to send `sessionId` for user specific operations.

### userPresent
```
let resp = await userPresent(username)
console.log(resp)
```

### userIsLoggedIn
```
let resp = await userIsLoggedIn(username)
console.log(resp)
```

### userLogout
```
let resp = await userLogout(sessionId)
console.log(resp)
```

### userDelete
```
let resp = await userDelete(sessionId)
console.log(resp)
```

### userStat
```
let resp = await userStat(sessionId)
console.log(resp)
```

### podNew
```
let resp = await podNew(sessionId, podName)
console.log(resp)
```

### podOpen
```
let resp = await podOpen(sessionId, podName)
console.log(resp)
```

### podClose
```
let resp = await podClose(sessionId, podName)
console.log(resp)
```

### podSync
```
let resp = await podSync(sessionId, podName)
console.log(resp)
```

### podDelete
```
let resp = await podDelete(sessionId, podName)
console.log(resp)
```

### podList
```
let resp = await podList(sessionId)
console.log(resp)
```

### podStat
```
let resp = await podStat(sessionId, podName)
console.log(resp)
```

### podShare
```
let resp = await podShare(sessionId, podName, shareAs)
console.log(resp)
```

### podReceive
```
let resp = await podReceive(sessionId, pod_sharing_reference)
console.log(resp)
```

### podReceiveInfo
```
let resp = await podReceiveInfo(sessionId, pod_sharing_reference)
console.log(resp)
```

### dirPresent
```
let resp = await dirPresent(sessionId, podName, dirPath)
console.log(resp)
```

### dirMake
```
let resp = await dirMake(sessionId, podName, dirPath)
console.log(resp)
```

### dirRemove
```
let resp = await dirRemove(sessionId, podName, dirPath)
console.log(resp)
```

### dirList
```
let resp = await dirList(sessionId, podName, dirPath)
console.log(resp)
```

### dirStat
```
let resp = await dirStat(sessionId, podName, dirPath)
console.log(resp)
```

### fileDownload
```
let resp = await fileDownload(sessionId, podName, filePath)
console.log(resp)
```

### fileUpload
```
let resp = await fileUpload(sessionId, podName, dirPath, fileByteArray, fileName, size, blockSize, compression)
console.log(resp)
```

### fileShare
```
let resp = await fileShare(sessionId, podName, dirPath, destinationUser)
console.log(resp)
```

### fileReceive
```
let resp = await fileReceive(sessionId, podName, directory, file_sharing_reference)
console.log(resp)
```

### fileReceiveInfo
```
let resp = await fileReceiveInfo(sessionId, podName, file_sharing_reference)
console.log(resp)
```

### fileDelete
```
let resp = await fileReceiveInfo(sessionId, podName, filePath)
console.log(resp)
```

### fileStat
```
let resp = await fileStat(sessionId, podName, filePath)
console.log(resp)
```

### kvNewStore
```
let resp = await kvNewStore(sessionId, podName, tableName, indexType)
console.log(resp)
```

Note: indexType is any of "string" or "number"

### kvList
```
let resp = await kvList(sessionId, podName)
console.log(resp)
```

### kvOpen
```
let resp = await kvOpen(sessionId, podName, tableName)
console.log(resp)
```

### kvDelete
```
let resp = await kvDelete(sessionId, podName, tableName)
console.log(resp)
```

### kvCount
```
let resp = await kvCount(sessionId, podName, tableName)
console.log(resp)
```

### kvEntryPut
```
let resp = await kvEntryPut(sessionId, podName, tableName, key, value)
console.log(resp)
```

### kvEntryGet
```
let resp = await kvEntryGet(sessionId, podName, tableName, key)
console.log(resp)
```

### kvEntryDelete
```
let resp = await kvEntryDelete(sessionId, podName, tableName, key)
console.log(resp)
```

### kvLoadCSV
```
let resp = await kvLoadCSV(sessionId, podName, tableName, memory, file)
console.log(resp)
```

### kvSeek
```
let resp = await kvSeek(sessionId, podName, tableName, start, end, limit)
```

### kvSeekNext
```
let resp = await kvSeekNext(sessionId, podName, tableName)
console.log(resp)
```

### docNewStore
```
let resp = await docNewStore(sessionId, podName, tableName, simpleIndexes, mutable)
console.log(resp)
```

Note: simpleIndexes is string_of_field=indexType pairs seperated with comma (first_name=string,age=number,tags=map)

### docList
```
let resp = await docList(sessionId, podName)
console.log(resp)
```

### docOpen
```
let resp = await docOpen(sessionId, podName, tableName)
console.log(resp)
```

### docCount
```
let resp = await docCount(sessionId, podName, tableName, expression)
console.log(resp)
```

### docDelete
```
let resp = await docDelete(sessionId, podName, tableName)
console.log(resp)
```

### docFind
```
let resp = await docFind(sessionId, podName, tableName, expression, limit)
console.log(resp)
```

### docEntryPut
```
let resp = await docEntryPut(sessionId, podName, tableName, value)
console.log(resp)
```

### docEntryGet
```
let resp = await docEntryGet(sessionId, podName, tableName, id)
console.log(resp)
```

### docEntryDelete
```
let resp = await docEntryDelete(sessionId, podName, tableName, id)
console.log(resp)
```

### docLoadJson
```
let resp = await docLoadJson(sessionId, podName, tableName, file)
console.log(resp)
```

### docIndexJson
```
let resp = await docIndexJson(sessionId, podName, tableName, filePath)
console.log(resp)
```



