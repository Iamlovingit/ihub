package midware

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"ihub/pkg/api"
	"ihub/pkg/config"
	"ihub/pkg/constants"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func InitMidwares(r *gin.Engine) error {
	midwareMap := map[string]gin.HandlerFunc{
		"log":     LoggerToFile(),
		"trace":   Trace(),
		"inout":   InOut(),
		"auth":    Auth(),
		"approve": Approve(),
	}
	for _, mw := range config.GetConfig().Midwares {
		if f, ok := midwareMap[mw.Midware]; ok {
			r.Use(f)
		} else {
			return fmt.Errorf("midware %s is not exist", mw.Midware)
		}
	}
	return nil
}

type LogFormatter struct {
	TimestampFormat string
	LevelDesc       []string
}

func (f *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.RFC3339
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	levelText := "INFO"
	if entry.Level != logrus.InfoLevel {
		levelText = f.LevelDesc[entry.Level]
	}

	b.WriteString(entry.Time.Format(timestampFormat))
	b.WriteByte(' ')
	b.WriteString(levelText)
	b.WriteByte(' ')
	b.WriteString(entry.Message)
	b.WriteByte('\n')

	return b.Bytes(), nil
}

// 日志中间件 将信息输出到标准输出和日志中
func LoggerToFile() gin.HandlerFunc {
	logger := logrus.New()
	logger.SetFormatter(&LogFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LevelDesc:       []string{"DEBUG", "INFO", "WARNING", "ERROR", "FATAL"},
	})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.DebugLevel)

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		c.Next()
		end := time.Now()
		latency := end.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		if raw != "" {
			path = path + "?" + raw
		}
		logger.WithFields(logrus.Fields{
			"status":     statusCode,
			"method":     method,
			"ip":         clientIP,
			"path":       path,
			"latency":    latency,
			"user-agent": c.Request.UserAgent(),
		}).Info()
	}
}

// Auth
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

// Approve
func Approve() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取运行模式
		runmode := config.GetConfig().Runmode
		// 获取用户角色
		role, exists := c.Get("Role")
		// 异常处理
		if !exists { // 如果用户角色获取失败
			rp := api.Reply{
				Code:    0,
				Message: "用户角色获取失败",
				Data:    "",
			}
			c.AbortWithStatusJSON(http.StatusOK, rp)
		} else if role == constants.RoleUser { // 如果为普通用户
			rp := api.Reply{
				Code:    0,
				Message: "普通用户无权限",
				Data:    "",
			}
			c.AbortWithStatusJSON(http.StatusOK, rp)
		}
		// 获取目的地
		destination, exists := c.Get("Destination")
		// 异常处理
		if !exists { // 如果目的地获取失败
			rp := api.Reply{
				Code:    0,
				Message: "目的地获取失败",
				Data:    "",
			}
			c.AbortWithStatusJSON(http.StatusOK, rp)
		} else if destination == constants.DestinationOut &&
			runmode == constants.RunmodeIn { // 集群外流量到达集群内网关(异常情况)
			rp := api.Reply{
				Code:    0,
				Message: "集群外流量不应到达集群内网关",
				Data:    "",
			}
			c.AbortWithStatusJSON(http.StatusOK, rp)
		}

		// [scheme:][//[userinfo@]host][/]path[?query][#fragment]
		// 解析URL，获取module和endpoint path = module/endpoint
		path := c.Request.URL.Path
		// 用 / 分割path
		pathSlice := strings.Split(path, "/")
		module := pathSlice[0]
		// 将除了第一个元素以外的元素拼接起来
		endpoint := strings.Join(pathSlice[1:], "/")

		// 如果以下两种字符串不为空，则说明是集群内流量
		// 从header 的 X-Cluster-Name 中获集群名称，或从args 的 clustername 中获取集群名称
		var clusterName string
		if c.Request.Header.Get("X-Cluster-Name") != "" ||
			c.Request.URL.Query().Get("clustername") != "" {
			// 获取集群名称
			if c.Request.Header.Get("X-Cluster-Name") != "" {
				clusterName = c.Request.Header.Get("X-Cluster-Name")
			} else {
				clusterName = c.Request.URL.Query().Get("clustername")
			}

			// 集群内流量到达集群外网关(Next交给Proxy处理)
			// 集群内流量到达集群内网关(需要审批?Insert:Next)
			// 集群外流量到达集群外网关(需要审批?Insert:Next)
			if destination == constants.DestinationIn &&
				runmode == constants.RunmodeOut { // 集群内流量到达集群外网关(Next交给Proxy处理)
				c.Next()
			} else if destination == constants.DestinationIn &&
				runmode == constants.RunmodeIn { // 集群内流量到达集群内网关(需要审批?Insert:Next)
				// 从Headers中获取信息
				// 获取用户ID
				//c.Request.Header.Get("X-User-Id")
				if role == constants.RoleClusterAdmin { // 如果为管理员
					needApprove, err := ClusterAdminNeedApprove(c.Request.Header, module, endpoint, clusterName)
					if err != nil {
						rp := api.Reply{
							Code:    0,
							Message: err.Error(),
							Data:    "",
						}
						c.AbortWithStatusJSON(http.StatusOK, rp)
					}
					c.Set("NeedApprove", needApprove)
					c.Next()
				} else if role == constants.RoleGroupAdmin { // 如果为组管理员
					// 获取组Id
					groupId, exists := c.Get("GroupId")
					if !exists {
						rp := api.Reply{
							Code:    0,
							Message: "组Id获取失败",
							Data:    "",
						}
						c.AbortWithStatusJSON(http.StatusOK, rp)
					}
					needApprove, err := GroupAdminNeedApprove(c.Request.Header, module, endpoint, clusterName, groupId.(int))
					if err != nil {
						rp := api.Reply{
							Code:    0,
							Message: err.Error(),
							Data:    "",
						}
						c.AbortWithStatusJSON(http.StatusOK, rp)
					}
					c.Set("NeedApprove", needApprove)
					c.Next()
				}
			} else if destination == constants.DestinationOut &&
				runmode == constants.RunmodeOut { // 集群外流量到达集群外网关(需要审批?Insert:Next)
				if role == constants.RoleClusterAdmin { // 如果为管理员
					needApprove, err := ClusterAdminNeedApprove(c.Request.Header, module, endpoint, clusterName)
					if err != nil {
						rp := api.Reply{
							Code:    0,
							Message: err.Error(),
							Data:    "",
						}
						c.AbortWithStatusJSON(http.StatusOK, rp)
					}
					c.Set("NeedApprove", needApprove)
					c.Next()
				} else if role == constants.RoleGroupAdmin { // 如果为组管理员
					// 获取组Id
					groupId, exists := c.Get("GroupId")
					if !exists {
						rp := api.Reply{
							Code:    0,
							Message: "组Id获取失败",
							Data:    "",
						}
						c.AbortWithStatusJSON(http.StatusOK, rp)
					}
					needApprove, err := GroupAdminNeedApprove(c.Request.Header, module, endpoint, clusterName, groupId.(int))
					if err != nil {
						rp := api.Reply{
							Code:    0,
							Message: err.Error(),
							Data:    "",
						}
						c.AbortWithStatusJSON(http.StatusOK, rp)
					}
					c.Set("NeedApprove", needApprove)
					c.Next()
				}
			}
		}
	}
}

func InOut() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
