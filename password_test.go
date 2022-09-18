package shuidiVPN


import "testing"

func TestRandPassword(t *testing.T) {
	passwd := RandPassword()
	t.Log("生产的随机密码：", passwd)
}

func TestParsePassword(t *testing.T) {
	passwd := RandPassword()
	t.Log("base64加密后的密码：", passwd)
	yuan_passwd, err := ParsePassword(passwd)
	if err != nil{
		t.Log("base64解析原密码失败")
		return
	}
	t.Log("base64解密后的密码：", yuan_passwd)
}