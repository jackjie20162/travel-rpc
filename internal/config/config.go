package config

import (
	"github.com/suyuan32/simple-admin-common/config"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DatabaseConf  config.DatabaseConf
	TranslateConf TranslateConf `json:",optional"`
}

// TranslateConf configures the external machine-translation provider used to
// fill the product/package translation tables (see 多语言多货币改造方案 P4).
// ApiKey must be supplied via the TRANSLATE_API_KEY environment variable and
// never hardcoded in yaml. When Provider/ApiKey are empty, auto-translation is
// disabled and the catalog simply serves base-language content.
type TranslateConf struct {
	Provider    string   `json:",optional"`              // 翻译提供方：deepl（可扩展）
	Endpoint    string   `json:",optional"`              // 翻译 API 端点
	ApiKey      string   `json:",optional"`              // 可选；缺省回退环境变量 TRANSLATE_API_KEY
	SourceLang  string   `json:",optional,default=auto"` // 源语言，auto 表示自动检测
	TargetLangs []string `json:",optional"`              // 目标语言（locale 码）：zh-CN/en-US/ar
}
