package db

import (
	"database/sql"
	"fmt"
	"hfproxy/pkg/config"
	"hfproxy/pkg/utils"
	"time"

	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
)

// ClusterManager .
type ClusterManager struct {
	ID             int            `db:"id"`
	Name           string         `db:"name"`
	Domain         string         `db:"domain"`
	Type           string         `db:"type"`
	SourceDivision sql.NullString `db:"source_division"`
	K8SVersion     sql.NullString `db:"k8s_version"`
	AISVersion     sql.NullString `db:"ais_version"`
	Description    sql.NullString `db:"description"`
	CreateTime     time.Time      `db:"creattime"`
	Flag           sql.NullString `db:"flag"`
	InstallPath    sql.NullString `db:"install_path"`
	Scan           sql.NullString `db:"scan"`
}

// ApproveRole .
type ApproveRole struct {
	ID        int    `db:"id"`
	UserID    int    `db:"user_id"`
	MOduleID  int    `db:"module_id"`
	Authority string `db:"authority"`
	ClusterID int    `db:"cluster_id"`
	UserRole  int    `db:"user_role"`
	GroupID   int    `db:"group_id"`
}

// ApproveModule .
type ApproveModule struct {
	ModuleID          int    `db:"module_id"`
	ModuleName        string `db:"module_name"`
	ModuleClusterType string `db:"module_cluster_type"`
	ModuleApprove     int    `db:"module_approve"`
}

// ApproveOperate .
type ApproveOperate struct {
	OperateID         int    `db:"operate_id"`
	OperateName       string `db:"operate_name"`
	OperateInterface  string `db:"operate_interface"`
	OperateAuthority  int    `db:"operate_authority"`
	ModuleID          int    `db:"module_id"`
	ModuleClusterType string `db:"module_cluster_type"`
}

// DBInstance global DB instance.
var DBInstance db.Session

// DBErr global DB error
var DBErr error

// 初始化数据库实例
func Init() error {
	dbcfg := config.GetConfig().DB
	var estr string
	fmt.Sscanf(dbcfg.Passwd, "%x", &estr)
	pwdByte, err := utils.DecryptSM2([]byte(estr), dbcfg.SM2PrivateFile)
	if err != nil {
		return err
	}
	var settings = mysql.ConnectionURL{
		Database: dbcfg.Name,
		Host:     fmt.Sprintf("%s:%d", dbcfg.Host, dbcfg.Port),
		User:     dbcfg.User,
		Password: string(pwdByte),
	}

	DBInstance, DBErr = mysql.Open(settings)
	if DBErr != nil {
		return DBErr
	}
	return nil
}

// GetDomainByClusterName .
// 根据集群名称获取对应的域名
func GetDomainByClusterName(cn string) (string, error) {
	fmt.Printf("cn: %v\n", cn)
	var cluster ClusterManager
	// 获取集群管理表的句柄
	c := DBInstance.Collection("cluster_manager")
	// 根据集群名称查询集群信息
	rec := c.Find(db.Cond{"name": cn})
	// 获取第一条记录并写入cluster变量中
	err := rec.One(&cluster)
	// 如果出错，打印错误信息并返回错误
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return "", err
	}
	fmt.Printf("cluster: %v\n", cluster)
	// 返回集群的域名
	return cluster.Domain, nil
}

func GetAuthority() ([]byte, error) {
	return []byte(""), nil
}

func GetDefaultAuthority() ([]byte, error) {
	var auth = make(map[string]int)
	var operator []ApproveOperate
	// 获取 approve_operate 表的句柄
	c := DBInstance.Collection("approve_operate")
	rec := c.Find(db.Cond{"operate_name": "default"})
	err := rec.All(operator)
	// 如果出错，打印错误信息并返回错误
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return []byte(""), err
	}
	return []byte(""), nil
}

func GetAuthorityByApproveRole(approveRole int, moduleName string, clusterID int, operateName string) (string, error) {
	// 如果 clusterID 为 -1 ，则返回错误

	// 如果 approveRole 为 0 （集群管理员）
	// 		如果 ApproveRole 查询到 1 条数据，则返回该数据的 authority 字段
	// 		如果 ApproveRole 查询到 <1 条数据，则在 ApproveOperate 表中查询默认审批标志位
	// 		如果 ApproveOperate 查询到 >1 条数据，则返回错误

	// 如果 approveRole 为 1 （组管理员）

	return "", nil
}
