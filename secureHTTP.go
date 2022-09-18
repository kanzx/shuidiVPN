package shuidiVPN

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

const (
	bufSize = 1024
)

type SecureHTTPConn struct {
	net.Conn
	Cipher *Cipher
}


//启动本地监听
func ListenLocal(localAddr *net.TCPAddr, handleConn func(localConn *SecureHTTPConn), cipher *Cipher) error {
	// 监听本机 localAddr
	listenLocal, err := net.ListenTCP("tcp", localAddr)
	if err != nil {
		log.Printf("监听本机 %s 失败：%s\n", localAddr, err)
		return err
	}

	defer listenLocal.Close()

	// for 循环等待所有请求连接的客户端(例如浏览器)
	for {
		userConn, err := listenLocal.AcceptTCP()
		if err != nil {
			log.Println("与发出HTTPTCP请求的客户端连接失败：", err)
			continue
		}
		// userConn 被关闭时直接清除所有数据 不管没有发送的数据
		userConn.SetLinger(0)
		go handleConn(&SecureHTTPConn{
			Conn: userConn,
			Cipher: cipher,
		})
	}
}

//与中转服务器建立连接
func DialServer(serverAddr *net.TCPAddr, cipher *Cipher) (*SecureHTTPConn, error) {
	//与中转服务器建立连接
	serverConn, err := net.DialTCP("tcp", nil, serverAddr)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("与中转服务器 %s 建立连接失败：%s", serverAddr, err))
	}
	return &SecureHTTPConn{
		Conn: serverConn,
		Cipher: cipher,
	}, nil
}

// 把放在bs里的数据加密后立即全部写入输出流
func (secureHTTPConn *SecureHTTPConn)EncodeWrite(bs []byte) (n int, err error) {
	secureHTTPConn.Cipher.encode(bs)
	//log.Println("加密后的数据：", bs)
	return secureHTTPConn.Write(bs)
}

// 从src中源源不断的读取原数据加密后写入到dst，直到src中没有数据可以再读取
func (secureHTTPConn *SecureHTTPConn) EncodeCopy(dst net.Conn) error {
	buf := make([]byte, bufSize)
	for {
		readCount, errRead := secureHTTPConn.Read(buf) //读取接收到的原数据
		if errRead != nil {
			if errRead != io.EOF {
				return errRead
			} else {
				return nil
			}
		}

		//log.Println("读取到的原数据：", buf[:readCount])

		if readCount > 0 {
			writeCount, errWrite := (&SecureHTTPConn{
				Conn: dst,
				Cipher: secureHTTPConn.Cipher,
			}).EncodeWrite(buf[0:readCount])  //将原数据加密
			if errWrite != nil {
				return errWrite
			}
			if readCount != writeCount {
				return io.ErrShortWrite
			}
		}
	}
}

// 从输入流里读取加密过的数据，解密后把原数据放到bs里
func (secureHTTPConn *SecureHTTPConn)DecodeRead(bs []byte) (n int, err error) {
	n, err = secureHTTPConn.Read(bs)
	if err != nil {
		return
	}

	//log.Println("读取的加密数据：", bs[:n])
	//log.Println(fmt.Sprintf("读取的加密数据字符串：%x\n", bs[:n]))

	secureHTTPConn.Cipher.decode(bs[:n])
	return
}

// 从src中源源不断的读取加密后的数据解密后写入到dst，直到src中没有数据可以再读取
func (secureHTTPConn *SecureHTTPConn) DecodeCopy(dst io.Writer) error {
	buf := make([]byte, bufSize)
	for {
		readCount, errRead := secureHTTPConn.DecodeRead(buf)  //解密后读取原数据
		if errRead != nil {
			if errRead != io.EOF {
				return errRead
			} else {
				return nil
			}
		}
		if readCount > 0 {
			writeCount, errWrite := dst.Write(buf[0:readCount]) //将原数据发送给目标
			if errWrite != nil {
				return errWrite
			}
			if readCount != writeCount {
				return io.ErrShortWrite
			}
		}
	}
}
