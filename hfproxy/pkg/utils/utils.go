package utils

import (
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
func MakeURL(domain, proxyPath string) (string, string) {
	// proxyPath = /完整路径(fullPath) = /模块名称(moudle)/真实路径(realPath)
	// 目标URL
	targetURL := "http://"
	// 去掉proxyPath中的第一个/，得到完整的路径
	fullPath := proxyPath[1:]
	// 截取fullPath中直到第一个/的字符串，该字符串即为模块名称
	moudle := fullPath[:strings.Index(fullPath, "/")]
	// 拼接目标URL，格式为http://模块名称.default.域名
	targetURL = targetURL + moudle + ".default." + domain
	// 截取fullPath中第一个/之后的字符串，该字符串为真实路径
	realPath := fullPath[strings.Index(fullPath, "/"):]
	// 返回目标URL和真实路径
	return targetURL, realPath
}
