debug_run:
	FYNEVIM_LOG_LEVEL=DEBUG go run ./...

run:
	go run ./...

clean:
	rm -rf .tmp
	rm -rf FyneVim.app

package: clean
	go run fyne.io/fyne/v2/cmd/fyne package
	codesign --force --deep --sign - FyneVim.app # ad-hoc signature
	# cp foo.provisionProfile NeoNV.app/Contents
	# mv NeoNV.app .tmp

install:
	go run fyne.io/fyne/v2/cmd/fyne install
