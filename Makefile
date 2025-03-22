debug_run:
	FYNEVIM_LOG_LEVEL=DEBUG go run ./... foo

run:
	go run ./...

clean:
	rm -rf .tmp
	rm -rf neovim.app

package: clean
	go run fyne.io/fyne/v2/cmd/fyne package
	codesign --force --deep --sign - neovim.app # ad-hoc signature
	# cp foo.provisionProfile NeoNV.app/Contents
	# mv NeoNV.app .tmp

install:
	go get ./...
	rm -rf /Applications/fynevim.app
	go run fyne.io/fyne/v2/cmd/fyne install
