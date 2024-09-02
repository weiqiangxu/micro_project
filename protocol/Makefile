#TAG := v$(shell git-cz version --patch)
TAG := "v0.02"

.PHONY: gen-go
gen-go: export GIT_EDITOR=true
gen-go: export EDITOR=true
gen-go: export GOPROXY=https://goproxy.cn  
gen-go:
	@sh build.sh
	@go mod tidy
	@git add .
	@git commit -m "gen go & update tag $(TAG)"
	@git push
	@git tag -f $(TAG)
	@git push origin -f $(TAG)