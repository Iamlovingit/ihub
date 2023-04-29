package handler

import (
	"ihub/pkg/api"
	mydb "ihub/pkg/db"
	"ihub/pkg/utils"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

// Health .
func Health(c *gin.Context) {
	rp := api.Reply{
		Code:    0,
		Message: "ok",
		Data:    "",
	}
	c.JSON(http.StatusOK, rp)
}

func Proxy(c *gin.Context) {
	// 从请求头中获取X-Cluster-Name，该请求头中包含了当前请求需要访问的集群名称。
	// 如果请求头中不包含X-Cluster-Name，则返回一个错误信息。
	v, ok := c.Request.Header["X-Cluster-Name"]
	if !ok {
		rp := api.Reply{
			Code:    1,
			Message: "缺少集群名称",
			Data:    "",
		}
		c.JSON(http.StatusOK, rp)
		return
	}

	// 根据集群名称获取对应的域名
	domain, err := mydb.GetDomainByClusterName(v[0])
	if err != nil {
		rp := api.Reply{
			Code:    1,
			Message: err.Error(),
			Data:    "",
		}
		c.JSON(http.StatusOK, rp)
		return
	}

	// 获取目标URL和真实路径
	// c.Param("proxyPath")获取请求路径中的proxyPath参数，如localhost:8080/api/v1/<proxyPath>
	// proxyPath = /完整路径(fullPath) = /模块名称(moudle)/真实路径(realPath)
	// 目标URL(targetURL)，格式为http://模块名称.default.域名

	targetURL, realPath := utils.MakeURL(domain, c.Param("proxyPath"))

	// Parse方法将字符串解析为URL结构体，并返回一个指向URL结构体的指针和一个错误值。
	remote, err := url.Parse(targetURL)
	if err != nil {
		rp := api.Reply{
			Code:    1,
			Message: err.Error(),
			Data:    "",
		}
		c.JSON(http.StatusOK, rp)
		return
	}

	// 创建一个httputil.ReverseProxy类型的代理对象，并设置其属性，将请求转发到目标URL
	// NewSingleHostReverseProxy的参数是一个指向URL结构体的指针，用于指定目标URL。
	proxy := httputil.NewSingleHostReverseProxy(remote)
	// Director属性是一个函数，该函数用于修改请求的属性，例如修改请求头、请求路径等。
	// 该函数的第一个参数是一个指向http.Request类型的指针，用于获取请求的属性。
	// 函数体中将请求头、请求路径等属性修改为目标URL的属性。
	proxy.Director = func(req *http.Request) {
		// Header属性是一个map，用于存储请求头(Headers)
		req.Header = c.Request.Header
		// Host属性是请求头中的Host字段，用于指定请求的主机名(模块名称.default.域名)，会解析为IP地址。
		req.Host = remote.Host
		// Scheme属性是请求头中的Scheme字段，用于指定请求的协议(http、https)
		req.URL.Scheme = remote.Scheme
		// URL.Host属性是请求头中的Host字段，用于指定请求的主机名(模块名称.default.域名)
		req.URL.Host = remote.Host
		// URL.Path属性是请求头中的Path字段，用于指定请求的路径。
		req.URL.Path = realPath
	}
	// ServeHTTP方法用于将请求转发到目标URL
	// 第一个参数是一个ResponseWriter类型的对象，用于将响应返回给客户端。
	// 第二个参数是一个指向http.Request类型的指针，用于获取请求的属性，传递给Director函数。
	proxy.ServeHTTP(c.Writer, c.Request)
}
