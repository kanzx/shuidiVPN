package shuidiVPN

import (
	"log"
	"net"
)

/**
	客户端浏览器 ———请求———> 中继客户端 ———加密———> 中转服务器 ———解密———> 目标服务器

	客户端浏览器 <———解密——— 中继客户端 <———加密——— 中转服务器 <———响应——— 目标服务器
 */

type ListenClient struct {
	Cipher     *Cipher		 //解密器
	ListenAddr *net.TCPAddr  //本地监听地址
	RemoteAddr *net.TCPAddr  //远程地址
}

func NewClient(clientAddr, serverAddr, password string) (*ListenClient, error) {
	passwd, err := ParsePassword(password)
	if err != nil {
		return nil, err
	}

	//本地监听的地址
	structListenAddr, err := net.ResolveTCPAddr("tcp", clientAddr)
	if err != nil {
		return nil, err
	}

	//远程中转服务器监听的地址
	structRemoteAddr, err := net.ResolveTCPAddr("tcp", serverAddr)
	if err != nil {
		return nil, err
	}
	return &ListenClient{
		Cipher:     NewCipher(passwd),
		ListenAddr: structListenAddr,
		RemoteAddr: structRemoteAddr,
	}, nil
}

// 本地端启动监听，接收来自本机浏览器的连接
func (local *ListenClient) Listen() error {
	return ListenLocal(local.ListenAddr, local.handleConn, local.Cipher)
}

// 处理HTTP请求，浏览器 ——HTTP请求——> 中继客户端 ———加密———> 中转服务器
func (local *ListenClient) handleConn(userConn *SecureHTTPConn)  {
	log.Println("发出HTTP请求的客户端：", userConn.RemoteAddr())
	defer userConn.Close()

	//与中转服务器建立连接
	serverConn, err := DialServer(local.RemoteAddr, local.Cipher)
	if err != nil {
		log.Println(err)
		return
	}

	defer serverConn.Close()

	// 进行转发
	go func() {
		err := serverConn.DecodeCopy(userConn)
		if err != nil {
			// 在 copy 的过程中可能会存在网络超时等 error 被 return，只要有一个发生了错误就退出本次工作
			userConn.Close()
			serverConn.Close()
		}
	}()

	// 从 userConn（客户端浏览器） 发送数据发送到 serverConn（中转服务器），这里因为处在翻墙阶段出现网络错误的概率更大
	userConn.EncodeCopy(serverConn)
}
