package utils

// AppId 定义应用Id
const AppId = 1

const AppModuleNumber = 1000
const ModuleErrNumber = 1000

// PubModuleId 公共错误空间的模块Id
const PubModuleId = 0

const (
	//Unauthorized 未授权
	Unauthorized = AppId*AppModuleNumber*ModuleErrNumber + PubModuleId*ModuleErrNumber + iota

	//ParameterParseError 参数解析错误
	ParameterParseError
)
