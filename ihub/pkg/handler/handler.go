package handler

import (
	"encoding/json"
	"ihub/pkg/api"
	"ihub/pkg/config"
	"ihub/pkg/constants"
	"ihub/pkg/db"
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
	// 判断是集群内还是集群外服务
	destination, _ := c.Get(constants.Destination)
	// 判断是否需要审批
	needApprove, _ := c.Get(constants.NeedApprove)
	runmode := config.GetConfig().Runmode

	clusterDomain, exists := c.Get(constants.ClusterDomain)
	if !exists {
		clusterDomain = ""
	}
	module := c.Param("module")
	endpoint := c.Param("endpoint")
	// 如果是应用商店接口，需要进行接口转换，如去掉v1/helm、v1/store等
	if _, ok := config.GetConfig().ApproveMap.AppstoreTransMap[endpoint]; ok {
		endpoint = config.GetConfig().ApproveMap.AppstoreTransMap[endpoint]
	}

	// 重构url
	targetURL, realPath := utils.MakeURL(
		destination.(string),
		clusterDomain.(string),
		module,
		endpoint,
		runmode,
	)

	if needApprove.(bool) && destination.(string) == runmode {
		// 添加审批
		// vali调用
		valiUrl := utils.MakeValiURL(targetURL, module, endpoint)
		req := c.Request
		method := req.Method
		remote, err := url.Parse(valiUrl)
		if err != nil {
			rp := api.Reply{
				Code:    1,
				Message: err.Error(),
				Data:    "",
			}
			c.JSON(http.StatusOK, rp)
			return
		}
		// req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = remote.Path
		res, err := (&http.Client{}).Do(req)
		if res == nil || err != nil {
			rp := api.Reply{
				Code:    1,
				Message: err.Error(),
				Data:    "",
			}
			c.JSON(http.StatusOK, rp)
			return
		}
		defer res.Body.Close() // ???

		// 读取响应体
		resourceDetail := map[string]interface{}{}
		err = json.NewDecoder(res.Body).Decode(&resourceDetail)
		//body, err := httputil.DumpResponse(res, true)
		if err != nil {
			rp := api.Reply{
				Code:    1,
				Message: err.Error(),
				Data:    "",
			}
			c.JSON(http.StatusOK, rp)
			return
		}

		// 从响应体中读取信息
		code := resourceDetail["code"].(int)
		data := resourceDetail["data"].(map[string]interface{})
		resourceInfo := data["resource_info"].(map[string]interface{})
		if code != 0 {
			rp := api.Reply{
				Code:    1,
				Message: "审批失败",
				Data:    "",
			}
			c.JSON(http.StatusOK, rp)
			return
		}

		var okList bool
		userid, ok := data["userid"]
		okList = ok
		userRole, ok := data["user_role"]
		okList = okList && ok
		groupid, ok := data["groupid"]
		okList = okList && ok
		approvetype, ok := data["type"]
		okList = okList && ok
		username, ok := data["username"]
		okList = okList && ok
		groupname, ok := data["groupname"]
		okList = okList && ok
		nsid, ok := data["nsid"]
		okList = okList && ok
		if !okList {
			rp := api.Reply{
				Code:    1,
				Message: "审批返回值缺少字段",
				Data:    "",
			}
			c.JSON(http.StatusOK, rp)
			return
		}

		xToken := map[string]map[string]interface{}{
			"user_info": {
				"user_id":   userid.(int),
				"nsid":      nsid.(string),
				"groupid":   groupid.(string),
				"username":  username.(string),
				"groupname": groupname.(string),
				"user_role": userRole.(string),
			},
		}
		// 转为json
		xTokenJson, _ := json.Marshal(xToken)
		resourceInfoJson, _ := json.Marshal(resourceInfo)
		resourceDetailJson, _ := json.Marshal(resourceDetail)
		clusterId, _ := c.Get(constants.ClusterId)
		approveRole, _ := c.Get(constants.ApproveRole)
		// json_to_str?
		db.InsertApproveInf(resourceInfoJson, resourceDetailJson, xTokenJson, userid.(int), module, endpoint, constants.ApproveStatusApproving, clusterId.(int), userRole.(int), targetURL, method, approveRole.(int), groupid.(int), approvetype.(string))

		rp := api.Reply{
			Code:    0,
			Message: "审批中",
			Data:    "",
		}
		c.JSON(http.StatusOK, rp)
	} else {
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

}
