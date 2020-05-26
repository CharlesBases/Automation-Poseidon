package core

import "charlesbases/Automation-Poseidon/demo/core/base"

type AutomationRequest struct {
	RequestInt    base.Int    `json:"request_int,omitempty"`
	RequestFloat  base.Float  `json:"request_float,omitempty"`
	RequestString base.String `json:"request_string,omitempty"`
}

type AutomationResponse struct {
	ResponseInt    base.Int    `json:"response_int,omitempty"`
	ResponseFloat  base.Float  `json:"response_float,omitempty"`
	ResponseString base.String `json:"response_string,omitempty"`
}
