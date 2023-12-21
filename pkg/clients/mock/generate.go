//go:build generate

package mock

//go:generate go run -mod=mod github.com/golang/mock/mockgen -package application -destination=./applications/mock.go -source=../applications/client.go ServiceClient -build_flags=-mod=mod
//go:generate go run -mod=mod github.com/golang/mock/mockgen -package projects -destination=./projects/mock.go -source=../projects/client.go ServiceClient -build_flags=-mod=mod
//go:generate go run -mod=mod github.com/golang/mock/mockgen -package cluster -destination=./cluster/mock.go -source=../cluster/client.go ServiceClient -build_flags=-mod=mod
