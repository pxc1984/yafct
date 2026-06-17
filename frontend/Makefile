.PHONY: mobile-release

mobile-release:
	pnpm run mobile:release
	mkdir -p dist
	cp android/app/build/outputs/apk/release/app-release.apk dist/yafct.apk
