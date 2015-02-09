GOARCH=arm GOOS=linux GOARM=5 go build -o builds/boxed_pi
GOARCH=amd64 GOOS=darwin go build -o builds/boxed_darwin
GOARCH=amd64 GOOS=linux go build -o builds/boxed_linux
