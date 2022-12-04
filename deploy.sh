GOARCH=amd64 GOOS=linux go build -o main main.go
zip function.zip main
aws lambda update-function-code --function-name parrot_bot --zip-file fileb://function.zip