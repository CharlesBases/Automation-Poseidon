//go:generate proto -file=$GOFILE -package=bei
package bll

type BeiService interface {
	Adds(request *AddsRequest) (response *AddsResponse)
}
