package bll

type Bei struct{}

func NewBeiService() BeiService {
	return new(Bei)
}
