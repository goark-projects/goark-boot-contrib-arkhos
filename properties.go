package gbcarkhos

const (
	// BeanNameContainer 是 Arkhos Servlet 容器 Bean 的稳定名称。
	BeanNameContainer = "goark.boot.arkhos.container"
	// BeanNameServer 是嵌入式 Arkhos HTTP 服务 Bean 的稳定名称。
	BeanNameServer = "goark.boot.arkhos.server"
)

const (
	// DefaultAddress 是嵌入式 Web 服务的默认监听地址。
	DefaultAddress = ":8080"
)

const (
	// PropertyServerAddress 设置嵌入式 Web 服务监听地址。
	PropertyServerAddress = "goark.web.server.address"
	// PropertyServerReadTimeout 设置完整请求读取超时。
	PropertyServerReadTimeout = "goark.web.server.read-timeout"
	// PropertyServerReadHeaderTimeout 设置请求头读取超时。
	PropertyServerReadHeaderTimeout = "goark.web.server.read-header-timeout"
	// PropertyServerWriteTimeout 设置响应写出超时。
	PropertyServerWriteTimeout = "goark.web.server.write-timeout"
	// PropertyServerIdleTimeout 设置 keep-alive 空闲超时。
	PropertyServerIdleTimeout = "goark.web.server.idle-timeout"
	// PropertyServerMaxHeaderBytes 设置请求头最大字节数。
	PropertyServerMaxHeaderBytes = "goark.web.server.max-header-bytes"
	// PropertyServerMaxRequestBodySize 设置整个 HTTP 请求体最大字节数。
	PropertyServerMaxRequestBodySize = "goark.web.server.max-request-body-size"
)

const (
	// PropertyFormMaxBodySize 设置 URL 编码表单体解析上限。
	PropertyFormMaxBodySize = "goark.web.servlet.form.max-body-size"
)

const (
	// PropertyMultipartLocation 设置 multipart 临时文件目录。
	PropertyMultipartLocation = "goark.web.servlet.multipart.location"
	// PropertyMultipartMaxFileSize 设置单个上传文件最大字节数。
	PropertyMultipartMaxFileSize = "goark.web.servlet.multipart.max-file-size"
	// PropertyMultipartMaxRequestSize 设置 multipart 请求体最大字节数。
	PropertyMultipartMaxRequestSize = "goark.web.servlet.multipart.max-request-size"
	// PropertyMultipartFileSizeThreshold 设置 multipart 落盘前内存阈值。
	PropertyMultipartFileSizeThreshold = "goark.web.servlet.multipart.file-size-threshold"
)

const (
	// PropertyAsyncTimeout 设置 Servlet 异步请求超时。
	PropertyAsyncTimeout = "goark.web.servlet.async.timeout"
)
