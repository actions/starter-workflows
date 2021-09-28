# This is a default distribution plug-in.
# Do not modify this file as your changes will be overwritten on next update.
# If you want customize installation, please make a copy.
DISTRO_NAME="Debian (bullseye)"

TARBALL_URL['aarch64']="https://github.com/termux/proot-distro/releases/download/v2.2.0/debian-aarch64-pd-v2.2.0.tar.xz"
TARBALL_SHA256['aarch64']="162ec58dd3cfd4e8924ad64e9d9fa4ee0b4ea7ddcfb62a0f6c542c6e6079b0fd"
TARBALL_URL['arm']="https://github.com/termux/proot-distro/releases/download/v2.2.0/debian-arm-pd-v2.2.0.tar.xz"
TARBALL_SHA256['arm']="4d907f0b596b5040fbf0fa41c9da5eea9049ff64bf2f54ddbd3ab0e317b16aa9"
TARBALL_URL['i686']="https://github.com/termux/proot-distro/releases/download/v2.2.0/debian-i686-pd-v2.2.0.tar.xz"
TARBALL_SHA256['i686']="357fcdd86b1680ce65bd43b2d8a127277f513ec1464ae70fbffe53a8952c6b03"
TARBALL_URL['x86_64']="https://github.com/termux/proot-distro/releases/download/v2.2.0/debian-x86_64-pd-v2.2.0.tar.xz"
TARBALL_SHA256['x86_64']="5ce7f65e089831b37d1cddeb67cfe4f3c487a507226b90535f420e13a37b9434"

distro_setup() {
	# Don't update gvfs-daemons and udisks2
	run_proot_cmd apt-mark hold gvfs-daemons udisks2
}
