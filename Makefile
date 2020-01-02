TARGET=bin/phumpkin
SRCS=$(shell find . -type f -name '*.go')


ALL: $(TARGET)


$(TARGET): $(SRCS) cmd/phumpkin/assets.go
	go build -o $(TARGET) ./cmd/phumpkin

cmd/phumpkin/assets.go: cmd/phumpkin/assets_gen.go frontend/dist/index.html
	go generate ./cmd/phumpkin

frontend/dist/index.html: frontend/node_modules $(shell find frontend -type f -name '*.vue')
	cd frontend && npm run build

frontend/node_modules: frontend/package.json frontend/package-lock.json
	cd frontend && npm install

static:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o $(TARGET) ./cmd/phumpkin/

clean:
	$(RM) -rf bin frontend/dist


.PHONY: clean