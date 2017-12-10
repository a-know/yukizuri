.PHONY: all

serve:
	./yukizuri.bin -addr=":8080" -logging=true

build:
	go-assets-builder --package=main templates/ > templates.go
	go build -o yukizuri.bin

assets:
	go-assets-builder --package=main templates/ > templates.go

deploy:
	go-assets-builder --package=main templates/ > templates.go
	GOOS=linux GOARCH=amd64 go build -o yukizuri.bin
	rsync -a --backup-dir=./.rsync_backup/$(LANG=C date +%Y%m%d%H%M%S) -e ssh ./* webapp:/var/www/yukizuri/app
