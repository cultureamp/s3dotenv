package = github.com/cultureamp/s3dotenv

.PHONY: release
release: s3dotenv-darwin-amd64.gz s3dotenv-linux-amd64.gz

%.gz: %
	gzip $<

%-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -o $@ $(package)

%-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o $@ $(package)
