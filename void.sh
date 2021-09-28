# This is a default distribution plug-in.
# Do not modify this file as your changes will be overwritten on next update.
# If you want customize installation, please make a copy.
DISTRO_NAME="Void Linux"

TARBALL_URL['aarch64']="https://github.com/termux/proot-distro/releases/download/v1.10.1/void-aarch64-pd-v1.10.1.tar.xz"
TARBALL_SHA256['aarch64']="c463c632e786afe084afe18686a28a2f4af738681e8aa0527e355cc23239fb6c"
TARBALL_URL['arm']="https://github.com/termux/proot-distro/releases/download/v1.10.1/void-arm-pd-v1.10.1.tar.xz"
TARBALL_SHA256['arm']="a24296d79b72b6f6750173e4113126f0ab45f79e29535d0c4c0f1d13bc2198d5"
TARBALL_URL['i686']="https://github.com/termux/proot-distro/releases/download/v1.10.1/void-i686-pd-v1.10.1.tar.xz"
TARBALL_SHA256['i686']="752d7338621ef4f93d0426506cb2be4b4a6a9695d06110f5da75e415bd80d761"
TARBALL_URL['x86_64']="https://github.com/termux/proot-distro/releases/download/v1.10.1/void-x86_64-pd-v1.10.1.tar.xz"
TARBALL_SHA256['x86_64']="3144bf082b9a16a809f758b18aa3d6769f9492051eb8000896f2139ea6873397"

distro_setup() {
	# Set default shell to bash.
	run_proot_cmd usermod --shell /bin/bash root
}
