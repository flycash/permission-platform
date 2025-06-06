# 初始化项目环境
.PHONY: setup
setup:
	@sh ./scripts/setup.sh

# 格式化代码
.PHONY: fmt
fmt:
	@goimports -l -w $$(find . -type f -name '*.go' -not -path "./.idea/*" -not -path "./**/ioc/wire_gen.go" -not -path "./**/ioc/wire.go")
	@gofumpt -l -w $$(find . -type f -name '*.go' -not -path "./.idea/*" -not -path "./**/ioc/wire_gen.go" -not -path "./**/ioc/wire.go")

# 清理项目依赖
.PHONY: tidy
tidy:
	@go mod tidy -v

.PHONY: check
check:
	@$(MAKE) --no-print-directory fmt
	@$(MAKE) --no-print-directory tidy

# 代码规范检查
.PHONY: lint
lint:
	@golangci-lint run -c ./scripts/lint/.golangci.yaml ./...

# 单元测试
.PHONY: ut
ut:
	@go test -race -shuffle=on -short -failfast -tags=unit -count=1 ./...

# 集成测试
.PHONY: e2e_up
e2e_up:
	@docker compose -p permission-platform -f scripts/test_docker_compose.yml up -d

.PHONY: e2e_down
e2e_down:
	@docker compose -p permission-platform -f scripts/test_docker_compose.yml down -v

.PHONY: e2e
e2e:
	@$(MAKE) e2e_down
	@$(MAKE) e2e_up
	@go test -race -shuffle=on -failfast -tags=e2e -count=1 ./...
	@$(MAKE) e2e_down

# 基准测试
.PHONY:	bench
bench:
	@go test -bench=. -benchmem  ./...

# 生成gRPC相关文件
.PHONY: grpc
grpc:
	@buf format -w api/proto
	@buf lint api/proto
	@buf generate api/proto

# 生成go代码
.PHONY: gen
gen:
	@go generate ./...

.PHONY: mysqldump
mysqldump:
	@docker exec -i permission-platform-mysql-1 mysqldump -u root -proot --databases permission > ./scripts/mysql/data.sql 2>/dev/null
	@cat ./scripts/mysql/database.sql >  ./scripts/mysql/init.sql
	@echo ""                          >> ./scripts/mysql/init.sql
	@cat ./scripts/mysql/data.sql    >> ./scripts/mysql/init.sql

.PHONY: run_platform_only
run_platform_only:
	@cd cmd/platform && export EGO_DEBUG=true && go run main.go --config=../../config/config.yaml

.PHONY: run_platform
run_platform:
	@$(MAKE) e2e_down
	@$(MAKE) e2e_up
	@sleep 15
	@cd cmd/platform && export EGO_DEBUG=true && go run main.go --config=../../config/config.yaml