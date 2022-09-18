package shuidiVPN

import (
	"os"
	"testing"
)

func clearConfigFile() {
	os.Remove(configPath)
}

func TestReadConfig(t *testing.T)  {
	clearConfigFile()
	// 读取默认配置
	config := Config{}
	config.ReadConfig()

	t.Log("读取的默认配置：", config)
}
