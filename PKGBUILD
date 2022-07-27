# Maintainer: shulhan <ms@kilabit.info>

pkgname=share-tools
pkgver=0.39.0.r30.gcf3761a
pkgrel=1

pkgdesc="Miscellaneous CLI tools: epoch, ini, xtrk"
arch=(x86_64)
url='https://github.com/shuLhan/share'
license=('BSD')

makedepends=(
	'go'
	'git'
)

provides=('share-tools')

source=(
	"$pkgname::git+https://github.com/shuLhan/share.git"
)
md5sums=(
	'SKIP'
)

pkgver() {
	cd "${pkgname}"
	git describe --long --tags | sed 's/^v//;s/\([^-]*-g\)/r\1/;s/-/./g'
}

prepare() {
	cd "${pkgname}"
}

build() {
	cd "${pkgname}"
	make
}

package() {
	cd "${pkgname}"
	install -Dm755 _bin/linux-amd64/epoch $pkgdir/usr/bin/epoch
	install -Dm755 _bin/linux-amd64/ini   $pkgdir/usr/bin/ini
	install -Dm755 _bin/linux-amd64/xtrk  $pkgdir/usr/bin/xtrk
	install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
}
