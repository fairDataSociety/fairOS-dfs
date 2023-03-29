[![Go Reference](https://pkg.go.dev/badge/github.com/fairdatasociety/fairOS-dfs@master/gomobile.svg)](https://pkg.go.dev/github.com/fairdatasociety/fairOS-dfs@master/gomobile)

# gomobile client for fairOS-dfs

Now fairOS-dfs can be used as android or

This package creates a global `dfs.API` instance, then saves user password in that instance after a successful login and will function for that user only.

### How to build for android (aar for android)?
```
$ go mod tidy
$ go get golang.org/x/mobile
$ go generate
```

### How to build for ios?
```
Coming soon
```

### How to use in go codebase?
```
Coming soon
```

### How to use in android?

#### API
```

    // isConnected checks if dfs.API is already initialised
    public static native boolean isConnected();

    // connect with a bee and initialise dfs.API
    public static native void connect(String beeEndpoint, String postageBlockId, String network, String rpc, long logLevel) throws Exception;
    
    public static native String loginUser(String username, String password) throws Exception;   
    public static native boolean isUserPresent(String username);    
    public static native boolean isUserLoggedIn();  
    public static native void logoutUser() throws Exception;   
    public static native String statUser() throws Exception;
    
    public static native String newPod(String podName) throws Exception;
    public static native String podOpen(String podName) throws Exception;
    public static native void podClose(String podName) throws Exception;
    public static native void podDelete(String podName) throws Exception;
    public static native void podSync(String podName) throws Exception;
    public static native String podList() throws Exception;
    public static native String podStat(String podName) throws Exception;
    public static native boolean isPodPresent(String podName);
    public static native String podShare(String podName) throws Exception;
    public static native String podReceive(String podSharingReference) throws Exception;
    public static native String podReceiveInfo(String podSharingReference) throws Exception;

    public static native String dirPresent(String podName, String dirPath) throws Exception;
    public static native String dirMake(String podName, String dirPath) throws Exception;
    public static native String dirRemove(String podName, String dirPath) throws Exception;
    public static native String dirList(String podName, String dirPath) throws Exception;
    public static native String dirStat(String podName, String dirPath) throws Exception;

    public static native String fileShare(String podName, String dirPath, String destinationUser) throws Exception;
    public static native String fileReceive(String podName, String directory, String fileSharingReference) throws Exception;
    public static native String fileReceiveInfo(String podName, String fileSharingReference) throws Exception;
    public static native void fileDelete(String podName, String filePath) throws Exception;
    public static native String fileStat(String podName, String filePath) throws Exception;
    public static native void fileUpload(String podName, String filePath, String dirPath, String compression, String blockSize, boolean overwrite) throws Exception;
    public static native void blobUpload(byte[] data, String podName, String filePath, String dirPath, String compression, long size, long blockSize, boolean overwrite) throws Exception;
    public static native byte[] fileDownload(String podName, String filePath) throws Exception;

    public static native String version();
    
```
***Document store and KV store documentation coming soon***

#### API flow

To use the fairOS-dfs in your android app you have to download the `fairos.aar` file from the downloads section  

Before calling any of the functions we need to connect with the swarm through a bee node and initialise fairos dfs API. 
For that we call `connect` function with the necessary parameters. 

We follow the normal flow of tasks after that, such as login, list user pods, opening a pod and list files and directories, upload or download content.

The function names are pretty much self-explanatory.

Check out this working [demo](https://github.com/fairDataSociety/fairOS-dfs-android-demo).

*** The code is not fully tested. Please use [fairOS-dfs](https://github.com/fairDataSociety/fairOS-dfs) for a better experience.


