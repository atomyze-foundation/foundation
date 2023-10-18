package proto

//go:generate protoc -I=. --go_out=. batch.proto
//go:generate protoc -I=. --go_out=. report.proto
//go:generate protoc -I=. --go_out=. locks.proto
//go:generate cp batch.pb.go ../test/integration/proto/batch.pb.go
