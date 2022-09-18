package shuidiVPN

import (
	"reflect"
	"testing"
)

func TestCipher(t *testing.T)  {
	passwd := RandPassword()
	t.Log("经过base64编码的密码：", passwd)

	p, _ := ParsePassword(passwd)
	t.Log("解码经过base64的密码：", p)

	//创建一个解码器
	cipher := NewCipher(p)

	// 假设原数据是 [0～255]
	org := make([]byte, PasswordLen)
	for i := 0; i < PasswordLen; i++ {
		org[i] = byte(i)
	}

	// 复制一份原数据到 tmp
	tmp := make([]byte, PasswordLen)
	copy(tmp, org)
	t.Log("原数据", tmp) //[0～255]

	// 加密 tmp
	cipher.encode(tmp)
	t.Log("原数据替换为随机生成的密码：", tmp) //将原数据转为随机生成的密码 RandPassword()

	// 解密 tmp
	cipher.decode(tmp)
	t.Log("将随机密码还原为原数据：", tmp) //将随机生成的密码转为原密码

	if !reflect.DeepEqual(org, tmp) {
		t.Error("解码编码数据后无法还原数据，数据不对应")
	}
}
