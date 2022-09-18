package shuidiVPN

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
)

/**
	客户端浏览器 ———请求———> 中继客户端 ———加密———> 中转服务器 ———解密———> 目标服务器

	客户端浏览器 <———解密——— 中继客户端 <———加密——— 中转服务器 <———响应——— 目标服务器
 */

type ListenServer struct {
	Cipher     *Cipher
	ListenAddr *net.TCPAddr
}

//创建一个中转服务端
// 1. 监听代理客户端发来经过加密的请求数据
// 2. 解密代理客户端请求的加密数据，获取到真正目标的地址
// 3. 代替客户端访问真正目标地址，将返回的数据加密后转发给客户端
func NewServer(serverAddr, password string) (*ListenServer, error) {
	passwd, err := ParsePassword(password)
	if err != nil {
		return nil, err
	}

	// 定义一个 Server端(中转服务器) 监听的地址(serverAddr)
	structListenAddr, err := net.ResolveTCPAddr("tcp", serverAddr)
	if err != nil {
		return nil, err
	}

	// 将获取的加密密码和 Server端 监听的地址返回给 ListenServer
	return &ListenServer{
		Cipher:     NewCipher(passwd),
		ListenAddr: structListenAddr,
	}, nil
}

// 运行服务端并且监听来自代理客户端的请求
func (listenServer *ListenServer) Listen() error {
	return ListenLocal(listenServer.ListenAddr, listenServer.handleConn, listenServer.Cipher)
}

// 处理请求
func (listenServer *ListenServer) handleConn(clientConn *SecureHTTPConn) {
	defer clientConn.Close()
	log.Println("连接中转服务器的客户端：", clientConn.RemoteAddr())

	// 用来存放客户端HTTP请求数据的缓冲区
	buf := make([]byte, 256)

	//解密数据后读取真实的HTTP请求数据
	n, err := clientConn.DecodeRead(buf)
	if err != nil {
		log.Println("从客户端中读取数据失败：", err)
		return
	}

	//log.Println("Server端读取到的解密数据：", buf[:n])
	//log.Println(fmt.Sprintf("Server端读取到的解密数据字符串：%x\n", buf[:n]))

	var method, URL, targetaddr string
	// 从解密的数据读入method，url
	fmt.Sscanf(string(buf[:bytes.IndexByte(buf[:], '\n')]), "%s%s", &method, &URL)
	hostPortURL, err := url.Parse(URL)
	if err != nil {
		log.Println(err)
		return
	}

	var targetAddr *net.TCPAddr

	// 如果方法是CONNECT，则为https协议
	if method == "CONNECT" {
		targetaddr = hostPortURL.Scheme + ":" + hostPortURL.Opaque
		targetAddr, err = net.ResolveTCPAddr("tcp", targetaddr)
		if err != nil {
			log.Println("获取真目标 TCPAddr 失败(net.ResolveTCPAddr 失败)：", err)
			return
		}
	} else { //否则为http协议
		// 如果host不带端口，则默认为80
		if strings.Index(hostPortURL.Host, ":") == -1 { //host不带端口， 默认80
			targetaddr = hostPortURL.Host + ":80"
			targetAddr, err = net.ResolveTCPAddr("tcp", targetaddr)
			if err != nil {
				log.Println("获取真目标 TCPAddr 失败(net.ResolveTCPAddr 失败)：", err)
				return
			}
		}
	}

	log.Println("真正目标URL：", targetAddr)


	//获得真正目标的host和port后，向目标服务器发起tcp连接
	targetConn, err := net.DialTCP("tcp", nil, targetAddr)
	if err != nil {
		log.Println(err)
		return
	} else {
		//响应连接成功
		//如果使用https协议，需先向客户端表示连接建立完毕
		if method == "CONNECT" {
			success := []byte("HTTP/1.1 200 Connection established\r\n\r\n")
			listenServer.Cipher.encode(success) //加密响应成功数据
			clientConn.Write(success)  //响应客户端连接成功
		} else {
			//如果使用http协议，需将从客户端得到的http请求转发给目标服务端
			targetConn.Write(buf[:n])
		}
	}

	// 进行转发  中转服务器 ———解密———> 目标服务器
	// 从 （代理客户端）clientConn 读取加密数据，解密后发送到 （目标服务器）targetConn
	go func() {
		err := clientConn.DecodeCopy(targetConn)
		if err != nil {
			// 在 copy 的过程中可能会存在网络超时等 error 被 return，只要有一个发生了错误就退出本次工作
			clientConn.Close()
			targetConn.Close()
		}
	}()

	//中继客户端 <———加密——— 中转服务器 <———响应——— 目标服务器
	// 从 targetConn 读取数据发送到 clientConn，这里因为处在翻墙阶段出现网络错误的概率更大
	(&SecureHTTPConn{
		Cipher: clientConn.Cipher,
		Conn: targetConn,
	}).EncodeCopy(clientConn)
}