package utils

import (
	"fmt"
	"ihub/pkg/config"
	"ihub/pkg/constants"
	"os"
	"strings"

	"github.com/tjfoc/gmsm/x509"
)

// DecryptSM2 .
func DecryptSM2(cipherText []byte, priFileName string) ([]byte, error) {
	//1.打开私钥问价读取私钥
	file, err := os.Open(priFileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	buf := make([]byte, fileInfo.Size())
	_, err = file.Read(buf)
	if err != nil {
		return nil, err
	}
	//2.将pem格式私钥文件解码并反序列话
	privateKeyFromPem, err := x509.ReadPrivateKeyFromPem(buf, nil)
	if err != nil {
		return nil, err
	}
	//3.解密
	planiText, err := privateKeyFromPem.DecryptAsn1(cipherText)
	if err != nil {
		return nil, err
	}
	return planiText, nil
}

// 拼接URL
func MakeURL(destination string, domain string, module string, endpoint string, runmode string) (string, string) {
	// 判断URL为集群内还是集群外
	if destination == constants.DestinationOut {
		// 	拼接集群外URL
		// http://localhost:port/endpoints
		targetURL := "http://localhost:"
		port := config.GetConfig().ApproveMap.OuterServicePortMap[module]
		targetURL = targetURL + fmt.Sprint(port) + "/" + endpoint
		realPath := "/" + endpoint
		return targetURL, realPath

	} else if destination == constants.DestinationIn {
		if runmode == constants.RunmodeIn {
			// 如果是集群内ihub服务
			// http://模块名称.default.域名/endpoints
			targetURL := "http://"
			realPath := "/" + endpoint
			targetURL = targetURL + module + ".default." + domain + realPath
			return targetURL, realPath
		} else {
			// 如果是集群外ihub服务，则发往集群内ihub服务
			// http://ihub.default.域名/module/endpoints
			targetURL := "http://" + "ihub" + ".default." + domain // 待修改
			realPath := "/" + module + "/" + endpoint
			targetURL = targetURL + realPath
			return targetURL, realPath
		}
	} else {
		return "", ""
	}
}

func MakeValiURL(url string, module string, endpoint string) string {
	vali_name := "vali_" + module + "_" + endpoint
	vali_url := url[:strings.LastIndex(url, "/")] + "/" + vali_name
	return vali_url
}
