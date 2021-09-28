# This is a default distribution plug-in.
# Do not modify this file as your changes will be overwritten on next update.
# If you want customize installation, please make a copy.
DISTRO_NAME="Ubuntu (hirsute)"

TARBALL_URL['aarch64']="https://github.com/termux/proot-distro/releases/download/v2.3.1/ubuntu-aarch64-pd-v2.3.1.tar.xz"
TARBALL_SHA256['aarch64']="599a0af87b110a9eab9f6f84b43243e497a73403397aeddb0d0b3cdb4ea54aa6"
TARBALL_URL['arm']="https://github.com/termux/proot-distro/releases/download/v2.3.1/ubuntu-arm-pd-v2.3.1.tar.xz"
TARBALL_SHA256['arm']="541415c3475bf1e15a1a4e91e2b1291410ed63a1a4b6d403f9096754d8f2bd74"
TARBALL_URL['x86_64']="https://github.com/termux/proot-distro/releases/download/v2.3.1/ubuntu-x86_64-pd-v2.3.1.tar.xz"
TARBALL_SHA256['x86_64']="c728976dcc66eed5ab4cb550c96b9d3169f7a46dd56736732b5eba1c48b6c58e"

distro_setup() {
	# Don't update gvfs-daemons and udisks2
	run_proot_cmd apt-mark hold gvfs-daemons udisks2
}
