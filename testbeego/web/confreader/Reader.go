package confreader

import "github.com/Unknwon/goconfig"

var language string
var cfg *goconfig.ConfigFile

//ReadConfFile __
func ReadConfFile(strFileName string) error {
	confIndex, err := goconfig.LoadConfigFile(strFileName)
	if err != nil {
		return err
	}

	cfg = confIndex

	return nil
}

//GetValue __
func GetValue(strKey string) string {
	if len(language) <= 0 {
		language = "zh-CN"
	}

	strValue, err := cfg.GetValue(language, strKey)
	if err != nil {
		return "配置文件无该数据, seesion:" + language + ", key:" + strKey
	}

	return strValue
}

//SetLanguage __
func SetLanguage(strLanguage string) {
	if len(strLanguage) > 0 {
		language = strLanguage
	}
}
