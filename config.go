package shuidiVPN

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
)

var (
	// 配置文件路径
	configPath string
	// 默认的配置文件名称
	configFilename string
	//随机生成的加密密码
	randPassword string

)

type Config struct {
	ServerAddr string `json:"server"`
	ClientAddr string `json:"client"`
	Password string	`json:"password"`
}

func init() {
	packagePath, _ := os.Getwd()
	configFilename = "config.json"
	configPath = path.Join(packagePath, configFilename)
	randPassword = RandPassword()
}

//读取配置
func (config *Config) ReadConfig() {
	if _, err := os.Stat(configPath); os.IsNotExist(err){
		log.Printf("配置文件 %s 不存在，创建默认配置\n", configPath)
		config := Config{
			ServerAddr: ":1992",
			ClientAddr: ":1080",
			Password: randPassword,
		}
		config.SaveConfig()
	}

	log.Printf("从路径 %s 中读取配置\n", configPath)
	file, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("打开配置文件 %s 出错:%s", configPath, err)
		return
	}
	defer file.Close()

	//创建Json编码器
	err = json.NewDecoder(file).Decode(config)
	if err != nil {
		log.Fatalf("JSON 配置文件格式不正确:\n%s", file.Name())
		return
	}
}

// 保存配置到配置文件
func (config *Config) SaveConfig() {
	configJson, _ := json.MarshalIndent(config, "", "	")
	err := ioutil.WriteFile(configPath, configJson, 0644)
	if err != nil {
		fmt.Errorf("保存配置到文件 %s 出错: %s", configPath, err)
	}
	log.Printf("保存配置到文件 %s 成功\n", configPath)
}