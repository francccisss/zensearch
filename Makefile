GO_FILES := $(wildcard *.go)
TS_FILES := $(wildcard *.ts)
GO_TESTFILES := $(wildcard *_test.go)

.PHONY: check go_build go_test npm_build

check: lint_check format_check sec_check
	@echo "Running checks"

ifdef GO_FILES
go_build:  
	go build -o zensearch 
	cd ./crawler/ && go build -o crawler-bin 
	cd ./search-engine/ && go build -o search_engine-bin
endif

ifdef TS_FILES
npm_build:  
	npm install ./database/
	npm install ./express-server/
	npm run build --prefix ./database/
	npm run build --prefix ./express-server/
endif

ifdef GO_TESTFILES
go_test: 
	@echo "Go Tests"
else 
go_test: 
	@echo "No files matches _test.go"

endif

	
.PHONY:  lint_check format_check sec_check

lint_check: 
	-staticcheck ./...

format_check: 
	-go fmt ./...

sec_check: 
	-gosec ./...
	


