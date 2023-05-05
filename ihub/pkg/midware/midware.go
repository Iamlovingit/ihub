package midware

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"ihub/pkg/api"
	"ihub/pkg/config"
	"ihub/pkg/constants"
	"ihub/pkg/db"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func InitMidwares(r *gin.Engine) error {
	midwareMap := map[string]gin.HandlerFunc{
		"log":     GinLogger(),
		"trace":   Trace(),
		"inout":   InOut(),
		"auth":    Auth(),
		"approve": Approve(),
	}
	for _, mw := range config.GetConfig().Midwares {
		if f, ok := midwareMap[mw.Midware]; ok {
			r.Use(f)
		} else {
			return fmt.Errorf("%s:%s", mw.Midware, "not exist.")
		}
	}
	return nil
}

// BodyWriter ..
type BodyWriter struct {
	gin.ResponseWriter
	bodyBuf *bytes.Buffer
}

func (w *BodyWriter) Write(b []byte) (int, error) {
	w.bodyBuf.Write(b)
	return w.ResponseWriter.Write(b)
}

/*
GinLogger is created for ginlog. It put the msg to stdout and logs.
*/
func GinLogger() gin.HandlerFunc {
	logger := logrus.New()
	f, _ := os.OpenFile(constants.DefaultLogName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	logger.SetOutput(io.MultiWriter(os.Stdout, f))
	logger.SetNoLock()

	logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	return func(c *gin.Context) {
		level, err := logrus.ParseLevel(config.GetConfig().LOG.Level)
		if err != nil {
			level = logrus.TraceLevel
		}
		logger.SetLevel(level)
		startTime := time.Now()
		reqMethod := c.Request.Method
		reqBody, _ := ioutil.ReadAll(c.Request.Body)
		reqURI := c.Request.RequestURI
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))
		clientIP := c.ClientIP()
		traceID := c.Request.Header.Get(constants.HTTPHeaderTraceID)

		bw := &BodyWriter{
			bodyBuf:        bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}
		c.Writer = bw
		c.Next()
		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)
		statusCode := c.Writer.Status()

		logger.WithFields(logrus.Fields{
			"status_code":  statusCode,
			"trace_id":     traceID,
			"latency_time": latencyTime,
			"client_ip":    clientIP,
			"req_method":   reqMethod,
			"req_uri":      reqURI,
		}).Info()

		logger.WithFields(logrus.Fields{
			"Type":      "Request",
			"ReqUri":    reqURI,
			"ReqHeader": c.Request.Header,
			"ReqBody":   string(reqBody),
		}).Trace()

		logger.WithFields(logrus.Fields{
			"Type":      "Response",
			"Status":    statusCode,
			"resHeader": c.Writer.Header(),
			"ResBody":   bw.bodyBuf.String(),
		}).Trace()
	}
}

// Auth
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		ibaseUrl := "http://" + config.GetConfig().IbaseUrl["ipPort"] + config.GetConfig().IbaseUrl["path"]
		remote, _ := url.Parse(ibaseUrl)
		req := http.Request{
			Method: http.MethodGet,
			Host:   remote.Host,
			URL:    remote,
			Header: c.Request.Header,
		}
		resp, err := http.DefaultClient.Do(&req)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			c.AbortWithError(http.StatusInternalServerError, errors.New("auth failed"))
			return
		}
		respData := map[string]interface{}{}
		json.NewDecoder(resp.Body).Decode(&respData)
		account := respData["account"].(string)
		groupId := respData["groupId"].(string)
		groupName := respData["groupName"].(string)
		roleType := respData["roleType"].(int)
		userId := respData["userId"].(string)
		userType := respData["userType"].(int)
		xAuthInfo := fmt.Sprintf("account:%s,groupId:%s,groupName:%s,roleType:%d,userId:%s,userType:%d",
			account, groupId, groupName, roleType, userId, userType)

		// p使用Base64对xAuthInfo进行编码
		pwdByte := base64.StdEncoding.EncodeToString([]byte(xAuthInfo))
		c.Request.Header.Set(constants.HTTPHeaderAuthInfo, string(pwdByte))
		c.Next()
	}
}

func inCheckList(module string, endpoint string) (bool, int) {
	// 如果 module 在 config.GetConfig().ApproveMap.ModuleOperateMapAdmin 的 keys 中，
	// 且 endpoint 在 config.GetConfig().ApproveMap.ModuleOperateMapAdmin[module] 列表中，
	// 则返回 true， 0
	if _, ok := config.GetConfig().ApproveMap.ModuleOperateMapAdmin[module]; ok {
		for _, v := range config.GetConfig().ApproveMap.ModuleOperateMapAdmin[module] {
			if v == endpoint {
				return true, 0
			}
		}
	} else if _, ok := config.GetConfig().ApproveMap.ModuleOperateMapGroup[module]; ok {
		for _, v := range config.GetConfig().ApproveMap.ModuleOperateMapGroup[module] {
			if v == endpoint {
				return true, 1
			}
		}
	}

	return false, 2
}

// Approve
func Approve() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取运行模式
		runmode := config.GetConfig().Runmode
		// // 获取用户角色
		// role, exists := c.Get("Role")

		// 获取目的地
		destination, exists := c.Get(constants.Destination)
		// 异常处理
		if !exists { // 如果目的地获取失败
			rp := api.Reply{
				Code:    999,
				Message: "目的地获取失败",
				Data:    "",
			}
			c.AbortWithStatusJSON(http.StatusOK, rp)
		} else if destination == constants.DestinationOut &&
			runmode == constants.RunmodeIn { // 集群外流量到达集群内网关(异常情况)
			rp := api.Reply{
				Code:    999,
				Message: "集群外流量不应到达集群内网关",
				Data:    "",
			}
			c.AbortWithStatusJSON(http.StatusOK, rp)
		}

		// [scheme:][//[userinfo@]host][/]path[?query][#fragment]
		// 解析URL，获取module和endpoint path = module/endpoint
		module := c.Param("moudle")

		endpoint := c.Param("endpoint")
		// 如果是应用商店接口，需要进行接口转换，如去掉v1/helm、v1/store等
		if _, ok := config.GetConfig().ApproveMap.AppstoreTransMap[endpoint]; ok {
			endpoint = config.GetConfig().ApproveMap.AppstoreTransMap[endpoint]
		}

		// 判断该模块/操作是否可能审批，若可能审批则返回需要审批的角色
		inList, role := inCheckList(module, endpoint)
		// 如果不在列表，则直接通过
		if !inList {
			c.Set(constants.NeedApprove, false)
			c.Next()
		}

		// 获取集群名、集群域名
		clusterName, exists := c.Get(constants.ClusterName)
		if !exists {
			rp := api.Reply{
				Code:    999,
				Message: "集群名获取失败",
				Data:    "",
			}
			c.AbortWithStatusJSON(http.StatusOK, rp)
		}
		// 集群内流量到达集群外网关(Next交给Proxy处理)
		// 集群内(外)流量到达集群内(外)网关(需要审批?Insert:Next)
		if destination == constants.DestinationIn &&
			runmode == constants.RunmodeOut { // 集群内流量到达集群外网关(Next交给Proxy处理)
			c.Next()
		} else {
			if role == constants.RoleClusterAdmin { // 如果为管理员
				needApprove, err := ClusterAdminNeedApprove(c.Request.Header, module, endpoint, clusterName.(string))
				if err != nil {
					rp := api.Reply{
						Code:    999,
						Message: err.Error(),
						Data:    "",
					}
					c.AbortWithStatusJSON(http.StatusOK, rp)
				}
				c.Set(constants.NeedApprove, needApprove)
				c.Set(constants.ApproveRole, constants.RoleClusterAdmin)
				c.Next()
			} else if role == constants.RoleGroupAdmin { // 如果为组管理员
				// 获取组Id
				groupId, exists := c.Get("GroupId")
				if !exists {
					rp := api.Reply{
						Code:    999,
						Message: "组Id获取失败",
						Data:    "",
					}
					c.AbortWithStatusJSON(http.StatusOK, rp)
				}
				needApprove, err := GroupAdminNeedApprove(c.Request.Header, module, endpoint, clusterName.(string), groupId.(int))
				if err != nil {
					rp := api.Reply{
						Code:    999,
						Message: err.Error(),
						Data:    "",
					}
					c.AbortWithStatusJSON(http.StatusOK, rp)
				}
				c.Set(constants.ApproveRole, constants.RoleGroupAdmin)
				c.Set(constants.NeedApprove, needApprove)
				c.Next()
			}
		}
	}
}

// 当域名数量为1时，判断集群异常状态
func checkClusterStatus(clusterName string) error {
	clusterStatus, err := db.GetClusterStatus(clusterName)
	if err != nil {
		return err
	}
	if clusterStatus == constants.ClusterStatusReseting || clusterStatus == constants.ClusterStatusResetSucceed {
		return errors.New("集群正在重置中")
	} else if clusterStatus == constants.ClusterStatusResetFailed {
		return errors.New("集群重置失败")
	} else {
		return nil
	}
}

// 检测域名合法性
func checkDomain(nameDomainId []db.NameDomainId) (api.Reply, error) {

	if len(nameDomainId) < 1 {
		rp := api.Reply{
			Code:    999,
			Message: "集群不存在",
			Data:    "",
		}
		return rp, errors.New("集群不存在")
	} else if len(nameDomainId) == 1 {
		// check
		clusterName := nameDomainId[0].Name
		err := checkClusterStatus(clusterName)
		if err != nil {
			rp := api.Reply{
				Code:    999,
				Message: err.Error(),
				Data:    "",
			}
			return rp, err
		}
		return api.Reply{}, nil
	} else {
		rp := api.Reply{
			Code:    999,
			Message: "集群Id获取失败",
			Data:    "",
		}
		return rp, errors.New("集群Id获取失败")
	}
}

func InOut() gin.HandlerFunc {
	return func(c *gin.Context) {
		// [scheme:][//[userinfo@]host][/]path[?query][#fragment]
		// 解析URL，获取module和endpoint path = module/endpoint
		module := c.Param("moudle")
		// 复制 c.Request.Header 和 c.Request.URL 防止被修改
		Request := c.Request.Clone(c)

		outerServiceMap := config.GetConfig().ApproveMap.OuterServicePortMap
		if _, ok := outerServiceMap[module]; ok { // 如果在集群外模块列表中
			c.Set(constants.Destination, constants.DestinationOut)
			c.Next()
		} else if Request.Header.Get("X-Cluster-Name") != "" ||
			Request.URL.Query().Get(constants.ClusterName) != "" { // 如果参数中有集群名称
			// 获取集群名称
			var clusterName string
			if Request.Header.Get("X-Cluster-Name") != "" {
				clusterName = Request.Header.Get("X-Cluster-Name")
			} else {
				clusterName = Request.URL.Query().Get(constants.ClusterName)
			}
			// 根据集群名获取集群域名
			nameDomainIdList, err := db.GetDomainIdByClusterName(clusterName)
			if err != nil {
				rp := api.Reply{
					Code:    999,
					Message: "集群不存在",
					Data:    "",
				}
				c.AbortWithStatusJSON(http.StatusOK, rp)
			}
			// 检测域名合法性
			rp, err := checkDomain(nameDomainIdList)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, rp)
			}
			// 设置集群名称、域名、目的地
			c.Set(constants.ClusterId, nameDomainIdList[0].ID)
			c.Set(constants.ClusterName, clusterName)
			c.Set(constants.ClusterDomain, nameDomainIdList[0].Domain)
			c.Set(constants.Destination, constants.DestinationIn)
			c.Next()
		} else if Request.Header.Get("X-Cluster-ID") != "" { // 如果Header中有集群Id
			// 获取集群Id
			clusterId, err := strconv.Atoi(Request.Header.Get("X-Cluster-ID"))
			if err != nil {
				rp := api.Reply{
					Code:    999,
					Message: "集群Id获取失败",
					Data:    "",
				}
				c.AbortWithStatusJSON(http.StatusOK, rp)
			}
			// 根据集群Id获取集群域名
			nameDomainIdList, err := db.GetNameDomainByClusterId(clusterId)
			if err != nil {
				rp := api.Reply{
					Code:    999,
					Message: err.Error(),
					Data:    "",
				}
				c.AbortWithStatusJSON(http.StatusOK, rp)
			}
			//checkDomain
			rp, err := checkDomain(nameDomainIdList)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, rp)
			}
			// 设置集群名称、域名、目的地
			c.Set(constants.ClusterName, nameDomainIdList[0].Name)
			c.Set(constants.ClusterId, nameDomainIdList[0].ID)
			c.Set(constants.ClusterDomain, nameDomainIdList[0].Domain)
			c.Set(constants.Destination, constants.DestinationIn)
			c.Next()
		} else {
			rp := api.Reply{
				Code:    999,
				Message: "路径不存在",
				Data:    "",
			}
			c.AbortWithStatusJSON(http.StatusNotFound, rp)
		}
	}
}

func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, ok := c.Request.Header[constants.HTTPHeaderTraceID]
		if !ok { //* X-Trace-ID is not exist, generate it.
			id := uuid.NewString()
			c.Request.Header.Set(constants.HTTPHeaderTraceID, id)
		}
	}
}
