package api

// Reply .
type Reply struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

//type TokenResponse struct {
//	ErrCode      string `json:"errCode"`
//	ErrMessage   string `json:"errMessage"`
//	ExceptionMsg string `json:"exceptionMsg"`
//	Flag         bool   `json:"flag"`
//	ResData      struct {
//		Account   string `json:"account"`
//		GroupId   string `json:"groupId"`
//		GroupName string `json:"groupName"`
//		InnerCall bool   `json:"innerCall"`
//		Ip        string `json:"ip"`
//		Priority  int    `json:"priority"`
//		RoleType  int    `json:"roleType"`
//		UserId    string `json:"userId"`
//		UserType  int    `json:"userType"`
//	} `json:"resData"`
//}
