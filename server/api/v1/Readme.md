# CROSS COMPILATION

* RASP
`env GOOS=linux GOARCH=arm GOARM=6 go build -v .`

* WINDOWS
`env GOOS=windows GOARCH=386 go build -v .`

* LINUX
`env GOOS=linux go build -v .`