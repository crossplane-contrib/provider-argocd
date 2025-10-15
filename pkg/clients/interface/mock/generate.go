//go:build generate

package mock

//go:generate go run -modfile ../../../../tools/go.mod -mod=mod go.uber.org/mock/mockgen -package application -destination=./applications/mock.go -source=../applications/client.go ServiceClient -build_flags=-mod=mod
//go:generate go run -modfile ../../../../tools/go.mod -mod=mod go.uber.org/mock/mockgen -package projects -destination=./projects/mock.go -source=../projects/client.go ServiceClient -build_flags=-mod=mod
//go:generate go run -modfile ../../../../tools/go.mod -mod=mod go.uber.org/mock/mockgen -package cluster -destination=./cluster/mock.go -source=../cluster/client.go ServiceClient -build_flags=-mod=mod
//go:generate go run -modfile ../../../../tools/go.mod -mod=mod go.uber.org/mock/mockgen -package applicationsets -destination=./applicationsets/mock.go -source=../applicationsets/client.go ServiceClient -build_flags=-mod=mod
//go:generate go run -modfile ../../../../tools/go.mod -mod=mod go.uber.org/mock/mockgen -package repositories -destination=./repositories/mock.go -source=../repositories/client.go ServiceClient -build_flags=-mod=mod
