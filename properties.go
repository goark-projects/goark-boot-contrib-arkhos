package gbcarkhos

const (
	// BeanNameContainer 是 Arkhos Servlet 容器 Bean 的稳定名称。
	BeanNameContainer = "goark.boot.arkhos.container"
	// BeanNameServer 是嵌入式 Arkhos HTTP 服务 Bean 的稳定名称。
	BeanNameServer = "goark.boot.arkhos.server"
	// BeanNameHertzLogger 是 Hertz 到 slog 的日志桥接 Bean 名称。
	BeanNameHertzLogger = "goark.boot.arkhos.hertzLogger"
)

const (
	// DefaultAddress 是嵌入式 Web 服务的默认监听地址。
	DefaultAddress = ":8080"
)

const (
	// PropertyServerAddress 设置监听主机。
	PropertyServerAddress = "server.address"
	// PropertyServerPort 设置监听端口。
	PropertyServerPort = "server.port"
	// PropertyServerShutdown 设置服务关闭模式，支持 immediate 和 graceful。
	PropertyServerShutdown = "server.shutdown"
	// PropertyServerMaxHTTPHeaderSize 设置 HTTP 请求头最大字节数。
	PropertyServerMaxHTTPHeaderSize = "server.max-http-request-header-size"
	// PropertyHertzReadTimeout 设置 Hertz 完整请求读取超时。
	PropertyHertzReadTimeout = "server.hertz.read-timeout"
	// PropertyHertzReadHeaderTimeout 设置 Hertz 请求头读取超时。
	PropertyHertzReadHeaderTimeout = "server.hertz.read-header-timeout"
	// PropertyHertzWriteTimeout 设置 Hertz 响应写出超时。
	PropertyHertzWriteTimeout = "server.hertz.write-timeout"
	// PropertyHertzIdleTimeout 设置 Hertz keep-alive 空闲超时。
	PropertyHertzIdleTimeout = "server.hertz.idle-timeout"
	// PropertyHertzMaxHeaderBytes 设置 Hertz 请求头最大字节数。
	PropertyHertzMaxHeaderBytes = "server.hertz.max-header-bytes"
	// PropertyHertzMaxRequestBodySize 设置 Hertz 请求体最大字节数。
	PropertyHertzMaxRequestBodySize = "server.hertz.max-request-body-size"
)

const (
	// PropertyMultipartEnabled 设置是否启用 multipart 解析。
	PropertyMultipartEnabled = "goark.servlet.multipart.enabled"
	// PropertyMultipartLocation 设置 multipart 临时文件目录。
	PropertyMultipartLocation = "goark.servlet.multipart.location"
	// PropertyMultipartMaxFileSize 设置单个上传文件最大字节数。
	PropertyMultipartMaxFileSize = "goark.servlet.multipart.max-file-size"
	// PropertyMultipartMaxRequestSize 设置 multipart 请求体最大字节数。
	PropertyMultipartMaxRequestSize = "goark.servlet.multipart.max-request-size"
	// PropertyMultipartFileSizeThreshold 设置 multipart 落盘前内存阈值。
	PropertyMultipartFileSizeThreshold = "goark.servlet.multipart.file-size-threshold"

	// PropertyFormMaxBodySize 设置 URL 编码表单体解析上限。
	PropertyFormMaxBodySize = "goark.servlet.form.max-body-size"
)

const (
	// PropertyAsyncTimeout 设置 Servlet 异步请求超时。
	PropertyAsyncTimeout = "goark.mvc.async.request-timeout"
)
