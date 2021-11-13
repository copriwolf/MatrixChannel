package util

import (
	"fmt"
	"io/ioutil"
	"matrixChannel/config"

	"gopkg.in/yaml.v3"
)

func GetConfig() (result *config.Config) {
	result = &config.Config{}
	yamlRes, err := ioutil.ReadFile("config/conf.yaml")
	if err != nil {
		panic(fmt.Sprintf("配置文件读取失败, err[%s]", err.Error()))
	}
	if err := yaml.Unmarshal(yamlRes, result); err != nil {
		panic(fmt.Sprintf("配置文件解析失败, err[%s]", err.Error()))
	}
	return
}

func SetConfig(conf *config.Config) {
	yamlRes, err := yaml.Marshal(conf)
	if err != nil {
		panic(fmt.Sprintf("配置文件序列化失败， err[%s]", err))
	}

	err = ioutil.WriteFile("conf.yaml", yamlRes, 0777)
	if err != nil {
		panic(fmt.Sprintf("配置文件写入失败, err[%s]", err.Error()))
	}
	return
}
