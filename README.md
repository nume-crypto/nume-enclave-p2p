# Nume enclave p2p

Enclave code for p2p protocol

## test

```sh
go test . -coverprofile coverage_temp.out -v
cat coverage_temp.out | grep -v "mimc.go\|merkle.go\|utils.go\|main.go" > coverage.out  
go tool cover -html=coverage.out
```
