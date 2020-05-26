package bll

import "charlesbases/Automation-Poseidon/demo/core"

type Automation interface {
	Add(request *core.AutomationRequest, requestInt int) (response *core.AutomationResponse, responseInt int, err error)
	Sub(request *core.AutomationRequest, requestBool bool) (response *core.AutomationResponse, responseBool bool, err error)
	Mul(request *core.AutomationRequest, requestFlost float64) (response *core.AutomationResponse, responseFlost float64, err error)
	Div(request *core.AutomationRequest, requestString string) (response *core.AutomationResponse, responseString string, err error)
}
