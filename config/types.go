package config

// Config 配置结构体
type Config struct {
	Server struct {
		Host         string `yaml:"host"`          // 服务器主机名
		Port         int    `yaml:"port"`          // 服务器端口
		InstanceName string `yaml:"instance_name"` // 实例名称
	} `yaml:"server"`

	SMTP struct {
		Hostname          string `yaml:"hostname"`            // SMTP 服务器主机名
		MaxSize           int    `yaml:"max_size"`            // 最大邮件大小
		MaxRecipients     int    `yaml:"max_recipients"`      // 最大收件人数量
		AuthFile          string `yaml:"auth_file"`           // 认证文件路径
		AllowAnonymous    bool   `yaml:"allow_anonymous"`     // 是否允许匿名访问
		AllowInsecureAuth bool   `yaml:"allow_insecure_auth"` // 是否允许不安全的认证
	} `yaml:"smtp"`

	Storage struct {
		Path string `yaml:"path"` // 存储路径
	} `yaml:"storage"`

	TLS struct {
		Enabled  bool   `yaml:"enabled"`   // 是否启用 TLS
		CertFile string `yaml:"cert_file"` // 证书文件路径
		KeyFile  string `yaml:"key_file"`  // 私钥文件路径
	} `yaml:"tls"`

	Log struct {
		Level     string `yaml:"level"`      // 日志级别：debug, info, warn, error
		Format    string `yaml:"format"`     // 日志格式：text, json
		File      string `yaml:"file"`       // 日志文件路径，为空则输出到标准输出
		AddSource bool   `yaml:"add_source"` // 是否添加源代码位置
	} `yaml:"log"`
}
