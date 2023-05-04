package midware

import (
	"fmt"
	"ihub/pkg/config"
	"ihub/pkg/constants"
	"ihub/pkg/db"
	"net/http"
)

func parseModulenameClusterid(module string, clusterName string) (string, int, error) {
	// 根据config将module映射为中英文字符串
	moduleName := config.GetConfig().ApproveMap.OperatorTransMap[module]
	// 获取集群域名和集群ID
	nameDomainIdList, err := db.GetDomainIdByClusterName(clusterName)
	if err != nil {
		return "", 0, err
	}
	clusterID := nameDomainIdList[0].ID
	// 如果集群ID为-1(默认集群)则返回错误
	if clusterID == -1 {
		return "", 0, fmt.Errorf("默认集群不需要审批")
	}
	return moduleName, clusterID, nil
}

// 查询默认审批权限
func defaultAuth(endpoint string, moduleName string, clusterID int) (bool, error) {
	// 根据配置文件及endpoint查询operatename
	operateName := config.GetConfig().ApproveMap.OperatorTransMap[endpoint]
	// 根据clusterId、Module、operateName查询默认审批权限
	DefaultauthorityList, err := db.GetDefaultauthority(clusterID, moduleName, operateName)
	if err != nil {
		return false, err
	}
	// 如果DefaultauthorityList中item数量为1,则根据authorityFlags查询是否需要审批
	// 其它情况报错
	if len(DefaultauthorityList) == 1 {
		// 根据authorityFlags查询是否需要审批
		return DefaultauthorityList[0].Authority == '1', nil
	} else {
		return false, fmt.Errorf("DefaultauthorityList item数量不为1")
	}
}

func ClusterAdminNeedApprove(header http.Header, module string, endpoint string, clusterName string) (bool, error) {

	// 解析信息
	moduleName, clusterID, err := parseModulenameClusterid(module, clusterName)
	if err != nil {
		return false, err
	}

	// 根据集群ID、模块名、role查询module_id和authority
	moduleAuthorityModuleIDList, err := db.GetModuleauthorityModuleid(clusterID, moduleName, constants.RoleClusterAdmin)
	// , err := db.GetModuleauthorityModuleid(clusterID, moduleName, constants.RoleClusterAdmin)
	if err != nil {
		return false, err
	}
	// 如果item数量为1,则根据moduleId和operateName查询是否需要审批
	// 如果item数量为0,则根据clusterId、Module、operateName查询默认审批权限
	// 其他情况报错
	if len(moduleAuthorityModuleIDList) == 1 {
		//
		authorityFlags := moduleAuthorityModuleIDList[0].Authority
		moduleId := moduleAuthorityModuleIDList[0].ModuleId
		// 根据operateName和moduleId查询operateId
		operatorIdList, err := db.GetOperatorid(moduleId, moduleName)
		if err != nil {
			return false, err
		}
		// 如果operatorIdList中item数量为1,则根据operatorid和authorityFlags查询是否需要审批
		// 其它情况报错
		if len(operatorIdList) == 1 {
			// 根据operatorId和authorityFlags查询是否需要审批
			operatorId := operatorIdList[0].OperatorId
			return authorityFlags[operatorId] == '1', nil
		} else {
			return false, fmt.Errorf("operatorIdList item数量不为1")
		}
	} else if len(moduleAuthorityModuleIDList) < 1 {
		return defaultAuth(endpoint, moduleName, clusterID)
	} else {
		return false, fmt.Errorf("moduleAuthorityModuleIDList item数量不为1")
	}
}

// 组管理员操作 判断是否需要审批
func GroupAdminNeedApprove(header http.Header, module string, endpoint string, clusterName string, groupId int) (bool, error) {
	// 解析信息
	moduleName, clusterID, err := parseModulenameClusterid(module, clusterName)
	if err != nil {
		return false, err
	}

	// 根据集群ID、模块名、groupId查询module_id和authority
	moduleAuthorityModuleIDList, err := db.GetModuleauthorityModuleid(clusterID, moduleName, groupId)
	if err != nil {
		return false, err
	}
	// 如果item数量为1,则根据moduleId和operateName查询是否需要审批
	// 如果item数量为0,则根据clusterId、Module、operateName查询默认审批权限
	// 其他情况报错
	if len(moduleAuthorityModuleIDList) == 1 {
		//
		authorityFlags := moduleAuthorityModuleIDList[0].Authority
		moduleId := moduleAuthorityModuleIDList[0].ModuleId
		// 根据operateName和moduleId查询operateId
		operatorIdList, err := db.GetOperatorid(moduleId, moduleName)
		if err != nil {
			return false, err
		}
		// 如果operatorIdList中item数量为1,则根据operatorid和authorityFlags查询是否需要审批
		// 其它情况报错
		if len(operatorIdList) == 1 {
			// 根据operatorId和authorityFlags查询是否需要审批
			operatorId := operatorIdList[0].OperatorId
			return authorityFlags[operatorId] == '1', nil
		} else {
			return false, fmt.Errorf("operatorIdList item数量不为1")
		}
	} else if len(moduleAuthorityModuleIDList) < 1 {
		return defaultAuth(endpoint, moduleName, clusterID)
	} else {
		return false, fmt.Errorf("moduleAuthorityModuleIDList item数量不为1")
	}
}
