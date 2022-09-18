# shuidiVPN

基于 [Lightsocks](https://github.com/gwuhaolin/lightsocks) 实现一个简单的 HTTP加密隧道，代理服务。

```json
	// 配置文件
	{
		"server": ":8888",
		"client": ":1080",
		"password": ""
	}
```

```go build 编译好 Server端 和 Client端```

在中转服务器运行 server端程序产生配置文件后，修改server监听地址 :8888 为自己需要的地址，再重新运行 server端

在客户端运行 Client端程序产生配置文件后，修改server监听地址 :8888 为中转服务器监听地址，修改client监听地址为自己客户端要监听的地址，再重新运行 client端