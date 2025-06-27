package router

type Result[T any] struct {
	Value T `json:"value"`
}
