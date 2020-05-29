package template

import "charlesbases/Automation-Poseidon/utils"

var poseidon = utils.Poseidon

type (
	Base struct{}

	ControllerInfor struct {
		InterfaceName string
		Func          *utils.Func
	}

	LogicInfor struct {
		Group string
	}
)
