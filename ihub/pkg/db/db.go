package db

import (
	"database/sql"
	"fmt"
	"ihub/pkg/config"
	"ihub/pkg/utils"
	"time"

	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
)

// ClusterManager .
type ClusterManager struct {
	ID             int            `db:"id, omitempty"`
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
	ID        int    `db:"id, omitempty"`
	UserID    int    `db:"user_id"`
	MOduleID  int    `db:"module_id"`
	Authority string `db:"authority"`
	ClusterID int    `db:"cluster_id"`
	UserRole  int    `db:"user_role"`
	GroupID   int    `db:"group_id"`
}

// ApproveModule .
type ApproveModule struct {
	ModuleID          int    `db:"module_id, omitempty"`
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

// ResetClusterTrace .
type ResetClusterTrace struct {
	ID          int       `db:"id, omitempty"`
	ClusterName string    `db:"cluster_name"`
	Status      string    `db:"status"`
	StepNum     int       `db:"step_num"`
	StepInfo    string    `db:"step_info"`
	TotalStep   int       `db:"total_step"`
	ErrorInfo   string    `db:"error_info"`
	InfoInfo    string    `db:"info_info"`
	ResetTime   time.Time `db:"reset_time"`
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

// DomainId .
type NameDomainId struct {
	Name   string `db:"name"`
	Domain string `db:"domain"`
	ID     int    `db:"id"`
}

// GetDomainIdByClusterName .
func GetDomainIdByClusterName(clusterName string) ([]NameDomainId, error) {
	var domainId []NameDomainId
	err := DBInstance.Collection("cluster_manager").
		Find(db.Cond{"name": clusterName}).
		All(&domainId)
	if err != nil {
		return nil, err
	}
	return domainId, nil
}

// GetDomainByClusterId .
func GetNameDomainByClusterId(clusterId int) ([]NameDomainId, error) {
	var nameDomain []NameDomainId
	err := DBInstance.Collection("cluster_manager").
		Find(db.Cond{"id": clusterId}).
		All(&nameDomain)
	if err != nil {
		return nil, err
	}
	return nameDomain, nil
}

// ClusterStatus .
type ClusterStatus struct {
	Status byte `db:"status"`
}

// GetClusterStatus .
func GetClusterStatus(clusterName string) (int, error) {
	var clusterStatus ClusterStatus
	err := DBInstance.Collection("reset_cluster_tracce").
		Find(db.Cond{"cluster_name": clusterName}).
		One(&clusterStatus)
	if err != nil {
		return 0, err
	}
	return int(clusterStatus.Status), nil
}

// ModuleauthorityModuleid .
type ModuleauthorityModuleid struct {
	Authority string `db:"authority"`
	ModuleId  int    `db:"module_id"`
}

// GetModuleauthorityModuleid .
// 根据clusterId、module、role查询module_id
func GetModuleauthorityModuleid(clusterId int, moduleName string, role int) ([]ModuleauthorityModuleid, error) {
	// 根据 module_id 连接 approve_rule、 approve_module 两张表查询
	var moduleauthorityModuleid []ModuleauthorityModuleid
	req := DBInstance.SQL().
		Select("a.authority, a.module_id").
		From("approve_role a").
		Join("approve_module b").
		On("a.module_id = b.module_id").
		Where(db.Cond{"a.cluster_id": clusterId, "b.module_name": moduleName, "a.user_role": role})
	err := req.All(&moduleauthorityModuleid)
	if err != nil {
		return nil, err
	}
	return moduleauthorityModuleid, nil
}

// GetModuleauthorityModuleidInGroup .
// 根据clusterId、module、组ID查询module_id
func GetModuleauthorityModuleidInGroup(clusterId int, moduleName string, groupId int) ([]ModuleauthorityModuleid, error) {
	// 根据 module_id 连接 approve_rule、 approve_module 两张表查询
	var moduleauthorityModuleid []ModuleauthorityModuleid
	req := DBInstance.SQL().
		Select("a.authority, a.module_id").
		From("approve_role a").
		Join("approve_module b").
		On("a.module_id = b.module_id").
		Where(db.Cond{"a.cluster_id": clusterId, "b.module_name": moduleName, "a.group_id": groupId})
	err := req.All(&moduleauthorityModuleid)
	if err != nil {
		return nil, err
	}
	return moduleauthorityModuleid, nil
}

// Defaultauthority .
type Defaultauthority struct {
	Authority byte `db:"authority"`
}

// GetDefaultauthority .
// 根据clusterId、module、operate_name查询默认权限
func GetDefaultauthority(clusterId int, moduleName string, operateName string) ([]Defaultauthority, error) {
	// 根据c.source_division b.module_cluster_type 连接 cluster_manager 和 approve_module 两张表查询
	// 根据 module_id 连接 approve_operate、 approve_module 两张表查询
	var defaultauthority []Defaultauthority
	req := DBInstance.SQL().
		Select("a.authority").
		From("approve_operate a").
		Join("approve_module b").
		On("a.module_id = b.module_id").
		Join("cluster_manager c").
		On("b.module_cluster_type = c.source_division OR b.module_cluster_type = 'All'").
		Where(db.Cond{"c.id": clusterId, "b.module_name": moduleName, "a.operate_name": operateName})
	err := req.All(&defaultauthority)
	if err != nil {
		return nil, err
	}
	return defaultauthority, nil
}

// Operatorid .
type Operatorid struct {
	OperatorId int `db:"operator_id"`
}

// GetOperatorid .
// 根据module_id、operate_name查询操操作id
func GetOperatorid(moduleId int, operateName string) ([]Operatorid, error) {
	// 根据 approve_operate 表查询
	var operatorid []Operatorid
	req := DBInstance.SQL().
		Select("operator_id").
		From("approve_operate").
		Where(db.Cond{"module_id": moduleId, "operate_name": operateName})
	err := req.All(&operatorid)
	if err != nil {
		return nil, err
	}
	return operatorid, nil
}

// ApproveInf .
type ApproveInf struct {
	ID              int       `db:"id, omitempty"`
	ResourceInfo    string    `db:"resource_info"`
	Createtime      time.Time `db:"createtime"`
	Userid          int       `db:"userid"`
	Type            string    `db:"type"`
	Status          string    `db:"status"`
	Cluasterid      int       `db:"clusterid"`
	Advice          string    `db:"advice"`
	Groupid         int       `db:"groupid"`
	UserRole        int       `db:"user_role"`
	ModuleName      string    `db:"module_name"`
	URL             string    `db:"url"`
	Method          string    `db:"method"`
	OperateName     string    `db:"operate_name"`
	ResourceDetail  string    `db:"resource_detail"`
	Headers         string    `db:"headers"`
	ApproveRole     int       `db:"approve_role"`
	ApproveResult   string    `db:"approve_result"`
	ApproveTime     time.Time `db:"approve_time"`
	IsDelete        int       `db:"is_delete"`
	IsApproveDelete int       `db:"is_approve_delete"`
}

// 向 approve_inf 插入一条记录
// resource_info resource_detail headers createtime userid module_name operate_name status clusterid
// user_role url method approve_role group_id type
// 其中 createtime 为当前时间， 使用SQL函数NOW()获取
func InsertApproveInf(resourceInfo []byte, resourceDetail []byte, headers []byte, userId int, moduleName string, operateName string, status string, clusterId int, userRole int, url string, method string, approveRole int, groupId int, approveType string) error {
	// 根据 approve_inf 表查询
	req := DBInstance.SQL().
		InsertInto("approve_inf").
		Columns("resource_info", "resource_detail", "headers", "createtime", "userid", "module_name", "operate_name", "status", "clusterid", "user_role", "url", "method", "approve_role", "groupid", "type").
		Values(resourceInfo, resourceDetail, headers, db.Raw("NOW()"), userId, moduleName, operateName, status, clusterId, userRole, url, method, approveRole, groupId, approveType)
	_, err := req.Exec()
	if err != nil {
		return err
	}
	return nil
}
