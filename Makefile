# https://encore.dev/guide/go.mod

#Чтобы превратить папку с кодом в Go-модуль, можно использовать команду go mod init:
init:
	go mod init pomodoro

# organize and clean up go.mod and go.sum (install|delete dependencies)
tidy:
	go mod tidy

# Run compiles and runs the named main Go package then deleted it
run:
	go run pomodoro

build:
	go build pomodoro

# format code
fmt:
	go fmt