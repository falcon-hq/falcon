New-Item -ItemType Directory -Force -Path bin
go build -ldflags "-s -w" -o bin/ .