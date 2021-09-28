#!@TERMUX_PREFIX@/bin/bash
##
## Script for managing proot'ed Linux distribution installations in Termux.
##
## Copyright (C) 2020-2021 Leonid Pliushch <leonid.pliushch@gmail.com>
##
## This program is free software: you can redistribute it and/or modify
## it under the terms of the GNU General Public License as published by
## the Free Software Foundation, either version 3 of the License, or
## (at your option) any later version.
##
## This program is distributed in the hope that it will be useful,
## but WITHOUT ANY WARRANTY; without even the implied warranty of
## MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
## GNU General Public License for more details.
##
## You should have received a copy of the GNU General Public License
## along with this program. If not, see <http://www.gnu.org/licenses/>.
##

PROGRAM_VERSION="2.6.2"

#############################################################################
#
# GLOBAL ENVIRONMENT AND INSTALLATION-SPECIFIC CONFIGURATION
#

set -e -u

PROGRAM_NAME="proot-distro"

# Where distribution plug-ins are stored.
DISTRO_PLUGINS_DIR="@TERMUX_PREFIX@/etc/proot-distro"

# Base directory where script keeps runtime data.
RUNTIME_DIR="@TERMUX_PREFIX@/var/lib/proot-distro"

# Where rootfs tarballs are downloaded.
DOWNLOAD_CACHE_DIR="${RUNTIME_DIR}/dlcache"

# Where extracted rootfs are stored.
INSTALLED_ROOTFS_DIR="${RUNTIME_DIR}/installed-rootfs"

# Colors.
if [ -n "$(command -v tput)" ] && [ $(tput colors) -ge 8 ] && [ -z "${PROOT_DISTRO_FORCE_NO_COLORS-}" ]; then
	RST="$(tput sgr0)"
	RED="${RST}$(tput setaf 1)"
	BRED="${RST}$(tput bold)$(tput setaf 1)"
	GREEN="${RST}$(tput setaf 2)"
	YELLOW="${RST}$(tput setaf 3)"
	BYELLOW="${RST}$(tput bold)$(tput setaf 3)"
	BLUE="${RST}$(tput setaf 4)"
	CYAN="${RST}$(tput setaf 6)"
	BCYAN="${RST}$(tput bold)$(tput setaf 6)"
	ICYAN="${RST}$(tput sitm)$(tput setaf 6)"
else
	RED=""
	BRED=""
	GREEN=""
	YELLOW=""
	BYELLOW=""
	BLUE=""
	CYAN=""
	BCYAN=""
	ICYAN=""
	RST=""
fi

#############################################################################
#
# FUNCTION TO PRINT A MESSAGE TO CONSOLE
#
# Prints a given text string to stderr. Handles escape sequences.
msg() {
	echo -e "$@" >&2
}

#############################################################################
#
# ANTI-ROOT FUSE
#
# This script should never be executed as root as can mess up the ownership,
# and SELinux labels in $PREFIX.
#
if [ "$(id -u)" = "0" ]; then
	msg
	msg "${BRED}Error: utility '${YELLOW}${PROGRAM_NAME}${BRED}' should not be used as root.${RST}"
	msg
	exit 1
fi

#############################################################################
#
# FUNCTION TO CHECK WHETHER DISTRIBUTION IS INSTALLED
#
# This is done by checking the presence of /bin directory in rootfs.
#
# Accepted arguments: $1 - name of distribution.
#
is_distro_installed() {
	if [ -e "${INSTALLED_ROOTFS_DIR}/${1}/bin" ]; then
		return 0
	else
		return 1
	fi
}

#############################################################################
#
# FUNCTION TO INSTALL THE SPECIFIED DISTRIBUTION
#
# Installs the Linux distribution by the following algorithm:
#
#  1. Checks whether requested distribution is supported, if yes - continue.
#  2. Checks whether requested distribution is installed, if not - continue.
#  3. Source the distribution configuration plug-in which contains the
#     functionality necessary for installation. It must define at least
#     TARBALL_URL and TARBALL_SHA256 associative array for at least one CPU
#     architecture.
#  4. Download the rootfs archive, if it is not available in cache.
#  5. Verify the rootfs archive if we have a SHA-256 for it. Otherwise print
#     a warning stating that integrity cannot be verified.
#  6. Extract the rootfs by using `tar` running under proot with link2symlink
#     extension. *Don't tell me that this is too slow.*
#  7. Write environment variables configuration to /etc/profile.d/termux-proot.sh.
#     If profile.d directory is not available, append to /etc/profile.
#  8. Write default /etc/resolv.conf.
#  9. Write default /etc/hosts.
#  10. Add missing Android specific UIDs/GIDs to user database.
#  11. Execute optional setup hook (distro_setup) if present.
#
command_install() {
	local distro_name
	local override_alias
	local distro_plugin_script

	while (($# >= 1)); do
		case "$1" in
			--)
				shift 1
				break
				;;
			--help)
				command_install_help
				return 0
				;;
			--override-alias)
				if [ $# -ge 2 ]; then
					shift 1

					if [ -z "$1" ]; then
						msg
						msg "${BRED}Error: argument to option '${YELLOW}--override-alias${BRED}' should not be empty.${RST}"
						command_install_help
						return 1
					fi

					if ! grep -qP '^[a-z0-9._+][a-z0-9._+-]+$' <<< "$1"; then
						msg
						msg "${BRED}Error: argument to option '${YELLOW}--override-alias${BRED}' should be lowercase and can contain only alphanumeric characters and these symbols '._+-'. Also argument should not begin with '-'.${RST}"
						msg
						return 1
					fi

					if grep -qP '^.*\.sh$' <<< "$1"; then
						msg
						msg "${BRED}Error: argument to option '${YELLOW}--override-alias${BRED}' should not end with '.sh'.${RST}"
						msg
						return 1
					fi

					override_alias="$1"
				else
					msg
					msg "${BRED}Error: option '${YELLOW}$1${BRED}' requires an argument.${RST}"
					command_install_help
					return 1
				fi
				;;
			-*)
				msg
				msg "${BRED}Error: unknown option '${YELLOW}${1}${BRED}'.${RST}"
				command_install_help
				return 1
				;;
			*)
				if [ -z "${distro_name-}" ]; then
					distro_name="$1"
				else
					msg
					msg "${BRED}Error: unknown option '${YELLOW}${1}${BRED}'.${RST}"
					msg
					msg "${BRED}Error: you have already set distribution as '${YELLOW}${distro_name}${BRED}'.${RST}"
					command_install_help
					return 1
				fi
				;;
		esac
		shift 1
	done

	if [ -z "${distro_name-}" ]; then
		msg
		msg "${BRED}Error: distribution alias is not specified.${RST}"
		command_install_help
		return 1
	fi

	if [ -z "${SUPPORTED_DISTRIBUTIONS["$distro_name"]+x}" ]; then
		msg
		msg "${BRED}Error: unknown distribution '${YELLOW}${distro_name}${BRED}' was requested to be installed.${RST}"
		msg
		msg "${CYAN}Run '${GREEN}${PROGRAM_NAME} list${CYAN}' to see the supported distributions.${RST}"
		msg
		return 1
	fi

	if [ -n "${override_alias-}" ]; then
		if [ ! -e "${DISTRO_PLUGINS_DIR}/${override_alias}.sh" ] && [ ! -e "${DISTRO_PLUGINS_DIR}/${override_alias}.override.sh" ]; then
			msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Creating file '${DISTRO_PLUGINS_DIR}/${override_alias}.override.sh'...${RST}"
			distro_plugin_script="${DISTRO_PLUGINS_DIR}/${override_alias}.override.sh"
			cp "${DISTRO_PLUGINS_DIR}/${distro_name}.sh" "${distro_plugin_script}"
			sed -i "s/^\(DISTRO_NAME=\)\(.*\)\$/\1\"${SUPPORTED_DISTRIBUTIONS["$distro_name"]} - ${override_alias}\"/g" "${distro_plugin_script}"
			SUPPORTED_DISTRIBUTIONS["${override_alias}"]="${SUPPORTED_DISTRIBUTIONS["$distro_name"]}"
			distro_name="${override_alias}"
		else
			msg
			msg "${BRED}Error: you cannot use value '${YELLOW}${override_alias}${BRED}' as alias override.${RST}"
			msg
			return 1
		fi
	else
		distro_plugin_script="${DISTRO_PLUGINS_DIR}/${distro_name}.sh"

		# Try an alternate distribution name.
		if [ ! -f "${distro_plugin_script}" ]; then
			distro_plugin_script="${DISTRO_PLUGINS_DIR}/${distro_name}.override.sh"
		fi
	fi

	if is_distro_installed "$distro_name"; then
		msg
		msg "${BRED}Error: distribution '${YELLOW}${distro_name}${BRED}' is already installed.${RST}"
		msg
		msg "${CYAN}Log in:     ${GREEN}${PROGRAM_NAME} login ${distro_name}${RST}"
		msg "${CYAN}Reinstall:  ${GREEN}${PROGRAM_NAME} reset ${distro_name}${RST}"
		msg "${CYAN}Uninstall:  ${GREEN}${PROGRAM_NAME} remove ${distro_name}${RST}"
		msg
		return 1
	fi

	if [ -f "${distro_plugin_script}" ]; then
		msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Installing ${YELLOW}${SUPPORTED_DISTRIBUTIONS["$distro_name"]}${CYAN}...${RST}"

		if [ ! -d "${INSTALLED_ROOTFS_DIR}/${distro_name}" ]; then
			msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Creating directory '${INSTALLED_ROOTFS_DIR}/${distro_name}'...${RST}"
			mkdir -m 755 -p "${INSTALLED_ROOTFS_DIR}/${distro_name}"
		fi

		if [ -d "${INSTALLED_ROOTFS_DIR}/${distro_name}/.l2s" ]; then
			export PROOT_L2S_DIR="${INSTALLED_ROOTFS_DIR}/${distro_name}/.l2s"
		fi

		# We need this to disable the preloaded libtermux-exec.so library
		# which redefines 'execve()' implementation.
		unset LD_PRELOAD

		# Needed for compatibility with some devices.
		#export PROOT_NO_SECCOMP=1

		# This should be overridden in distro plug-in with valid URL for
		# each architecture where possible.
		TARBALL_URL["aarch64"]=""
		TARBALL_URL["arm"]=""
		TARBALL_URL["i686"]=""
		TARBALL_URL["x86_64"]=""

		# This should be overridden in distro plug-in with valid SHA-256
		# for corresponding tarballs.
		TARBALL_SHA256["aarch64"]=""
		TARBALL_SHA256["arm"]=""
		TARBALL_SHA256["i686"]=""
		TARBALL_SHA256["x86_64"]=""

		# If your content inside tarball isn't stored in subdirectory,
		# you can override this variable in distro plug-in with 0.
		TARBALL_STRIP_OPT=1

		# Distribution plug-in contains steps on how to get download URL
		# and further post-installation configuration.
		source "${distro_plugin_script}"

		# Cannot proceed without URL and SHA-256.
		if [ -z "${TARBALL_URL["$DISTRO_ARCH"]}" ]; then
			msg "${BLUE}[${RED}!${BLUE}] ${CYAN}Sorry, but distribution download URL is not defined for CPU architecture '$DISTRO_ARCH'.${RST}"
			return 1
		fi
		if ! grep -qP '^[0-9a-fA-F]+$' <<< "${TARBALL_SHA256["$DISTRO_ARCH"]}"; then
			msg
			msg "${BRED}Error: got malformed SHA-256 from ${distro_plugin_script}${RST}"
			msg
			return 1
		fi

		if [ ! -d "$DOWNLOAD_CACHE_DIR" ]; then
			msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Creating directory '$DOWNLOAD_CACHE_DIR'...${RST}"
			mkdir -p "$DOWNLOAD_CACHE_DIR"
		fi

		local tarball_name
		tarball_name=$(basename "${TARBALL_URL["$DISTRO_ARCH"]}")

		if [ ! -f "${DOWNLOAD_CACHE_DIR}/${tarball_name}" ]; then
			msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Downloading rootfs tarball...${RST}"

			# Using temporary file as script can't distinguish the partially
			# downloaded file from the complete. Useful in case if curl will
			# fail for some reason.
			msg
			rm -f "${DOWNLOAD_CACHE_DIR}/${tarball_name}.tmp"
			if ! curl --fail --retry 5 --retry-connrefused --retry-delay 5 --location \
				--output "${DOWNLOAD_CACHE_DIR}/${tarball_name}.tmp" "${TARBALL_URL["$DISTRO_ARCH"]}"; then
				msg "${BLUE}[${RED}!${BLUE}] ${CYAN}Download failure, please check your network connection.${RST}"
				rm -f "${DOWNLOAD_CACHE_DIR}/${tarball_name}.tmp"
				return 1
			fi
			msg

			# If curl finished successfully, rename file to original.
			mv -f "${DOWNLOAD_CACHE_DIR}/${tarball_name}.tmp" "${DOWNLOAD_CACHE_DIR}/${tarball_name}"
		else
			msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Using cached rootfs tarball...${RST}"
		fi

		if [ -n "${TARBALL_SHA256["$DISTRO_ARCH"]}" ]; then
			msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Checking integrity, please wait...${RST}"
			local actual_sha256
			actual_sha256=$(sha256sum "${DOWNLOAD_CACHE_DIR}/${tarball_name}" | awk '{ print $1}')

			if [ "${TARBALL_SHA256["$DISTRO_ARCH"]}" != "${actual_sha256}" ]; then
				msg "${BLUE}[${RED}!${BLUE}] ${CYAN}Integrity checking failed. Try to redo installation again.${RST}"
				rm -f "${DOWNLOAD_CACHE_DIR}/${tarball_name}"
				return 1
			fi
		else
			msg "${BLUE}[${RED}!${BLUE}] ${CYAN}Integrity checking of downloaded rootfs has been disabled.${RST}"
		fi

		msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Extracting rootfs, please wait...${RST}"
		# --exclude='dev'||: - need to exclude /dev directory which may contain device files.
		# --delay-directory-restore - set directory permissions only when files were extracted
		#                             to avoid issues with Arch Linux bootstrap archives.
		proot --link2symlink \
			tar -C "${INSTALLED_ROOTFS_DIR}/${distro_name}" --warning=no-unknown-keyword \
			--delay-directory-restore --preserve-permissions --strip="$TARBALL_STRIP_OPT" \
			-xf "${DOWNLOAD_CACHE_DIR}/${tarball_name}" --exclude='dev'||:

		# Write important environment variables to profile file as /bin/login does not
		# preserve them.
		local profile_script
		if [ -d "${INSTALLED_ROOTFS_DIR}/${distro_name}/etc/profile.d" ]; then
			profile_script="${INSTALLED_ROOTFS_DIR}/${distro_name}/etc/profile.d/termux-proot.sh"
		else
			chmod u+rw "${INSTALLED_ROOTFS_DIR}/${distro_name}/etc/profile" >/dev/null 2>&1 || true
			profile_script="${INSTALLED_ROOTFS_DIR}/${distro_name}/etc/profile"
		fi
		msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Writing '$profile_script'...${RST}"
		local LIBGCC_S_PATH
		LIBGCC_S_PATH="/$(cd ${INSTALLED_ROOTFS_DIR}/${distro_name}; find usr/lib/ -name libgcc_s.so.1)"
		cat <<- EOF >> "$profile_script"
		export ANDROID_ART_ROOT=${ANDROID_ART_ROOT-}
		export ANDROID_DATA=${ANDROID_DATA-}
		export ANDROID_I18N_ROOT=${ANDROID_I18N_ROOT-}
		export ANDROID_ROOT=${ANDROID_ROOT-}
		export ANDROID_RUNTIME_ROOT=${ANDROID_RUNTIME_ROOT-}
		export ANDROID_TZDATA_ROOT=${ANDROID_TZDATA_ROOT-}
		export BOOTCLASSPATH=${BOOTCLASSPATH-}
		export COLORTERM=${COLORTERM-}
		export DEX2OATBOOTCLASSPATH=${DEX2OATBOOTCLASSPATH-}
		export EXTERNAL_STORAGE=${EXTERNAL_STORAGE-}
		export LANG=C.UTF-8
		export PATH=\${PATH}:/data/data/com.termux/files/usr/bin:/system/bin:/system/xbin
		export PREFIX=${PREFIX-/data/data/com.termux/files/usr}
		export TERM=${TERM-xterm-256color}
		export TMPDIR=/tmp
		export PULSE_SERVER=127.0.0.1
		export MOZ_FAKE_NO_SANDBOX=1
		EOF
		if [ "${LIBGCC_S_PATH}" != "/" ]; then
			echo "${LIBGCC_S_PATH}" >> "${INSTALLED_ROOTFS_DIR}/${distro_name}/etc/ld.so.preload"
			chmod 644 "${INSTALLED_ROOTFS_DIR}/${distro_name}/etc/ld.so.preload"
		fi
		unset LIBGCC_S_PATH

		# /etc/resolv.conf may not be configured, so write in it our configuraton.
		msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Writing resolv.conf file (NS 1.1.1.1/1.0.0.1)...${RST}"
		rm -f "${INSTALLED_ROOTFS_DIR}/${distro_name}/etc/resolv.conf"
		cat <<- EOF > "${INSTALLED_ROOTFS_DIR}/${distro_name}/etc/resolv.conf"
		nameserver 1.1.1.1
		nameserver 1.0.0.1
		EOF

		# /etc/hosts may be empty by default on some distributions.
		msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Writing hosts file...${RST}"
		chmod u+rw "${INSTALLED_ROOTFS_DIR}/${distro_name}/etc/hosts" >/dev/null 2>&1 || true
		cat <<- EOF > "${INSTALLED_ROOTFS_DIR}/${distro_name}/etc/hosts"
		# IPv4.
		127.0.0.1   localhost.localdomain localhost

		# IPv6.
		::1         localhost.localdomain localhost ip6-localhost ip6-loopback
		fe00::0     ip6-localnet
		ff00::0     ip6-mcastprefix
		ff02::1     ip6-allnodes
		ff02::2     ip6-allrouters
		ff02::3     ip6-allhosts
		EOF

		# Add Android-specific UIDs/GIDs to /etc/group and /etc/gshadow.
		msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Registering Android-specific UIDs and GIDs...${RST}"
		chmod u+rw "${INSTALLED_ROOTFS_DIR}/${distro_name}/etc/passwd" \
			"${INSTALLED_ROOTFS_DIR}/${distro_name}/etc/shadow" \
			"${INSTALLED_ROOTFS_DIR}/${distro_name}/etc/group" \
			"${INSTALLED_ROOTFS_DIR}/${distro_name}/etc/gshadow" >/dev/null 2>&1 || true
		echo "aid_$(id -un):x:$(id -u):$(id -g):Android user:/:/sbin/nologin" >> \
			"${INSTALLED_ROOTFS_DIR}/${distro_name}/etc/passwd"
		echo "aid_$(id -un):*:18446:0:99999:7:::" >> \
			"${INSTALLED_ROOTFS_DIR}/${distro_name}/etc/shadow"
		local group_name group_id
		while read -r group_name group_id; do
			echo "aid_${group_name}:x:${group_id}:root,aid_$(id -un)" >> \
				"${INSTALLED_ROOTFS_DIR}/${distro_name}/etc/group"
			if [ -f "${INSTALLED_ROOTFS_DIR}/${distro_name}/etc/gshadow" ]; then
				echo "aid_${group_name}:*::root,aid_$(id -un)" >> \
					"${INSTALLED_ROOTFS_DIR}/${distro_name}/etc/gshadow"
			fi
		done < <(paste <(id -Gn | tr ' ' '\n') <(id -G | tr ' ' '\n'))

		# Ensure that proot will be able to bind fake /proc entries.
		setup_fake_proc

		# Run optional distro-specific hook.
		if declare -f -F distro_setup >/dev/null 2>&1; then
			msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Running distro-specific configuration steps...${RST}"
			(cd "${INSTALLED_ROOTFS_DIR}/${distro_name}"
				distro_setup
			)
		fi

		msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Installation finished.${RST}"
		msg
		msg "${CYAN}Now run '${GREEN}$PROGRAM_NAME login $distro_name${CYAN}' to log in.${RST}"
		msg
		return 0
	else
		msg "${BLUE}[${RED}!${BLUE}] ${CYAN}Cannot find '${distro_plugin_script}' which contains distro-specific install functions.${RST}"
		return 1
	fi
}

# Special function for executing a command in rootfs.
# Can be used only inside distro_setup().
run_proot_cmd() {
	if [ -z "${distro_name-}" ]; then
		msg
		msg "${BRED}Error: called run_proot_cmd() but \${distro_name} is not set. Possible cause: using run_proot_cmd() outside of distro_setup()?${RST}"
		msg
		return 1
	fi

	if [ -z "${DISTRO_ARCH-}" ]; then
		msg
		msg "${BRED}Error: called run_proot_cmd() but \${DISTRO_ARCH} is not set.${RST}"
		msg
		return 1
	fi

	local qemu_arg=""
	if [ "$DISTRO_ARCH" != "$DEVICE_CPU_ARCH" ]; then
		if ! [[ "$DISTRO_ARCH" =~ ^(aarch64|arm|i686|x86_64)$ ]]; then
			msg
			msg "${BRED}Error: DISTRO_ARCH has unknown value '$target_arch'. Valid values are: aarch64, arm, i686, x86_64."
			msg
			return 1
		fi

		# If CPU and host OS are 64bit, we can run 32bit guest OS without emulation.
		# Everything else requires emulator (QEMU).
		if ! [ "$DEVICE_CPU_ARCH" = "aarch64" ] && [ "$DISTRO_ARCH" = "arm" ] || \
			! [ "$DEVICE_CPU_ARCH" = "x86_64" ] && [ "$DISTRO_ARCH" = "i686" ]; then

			if [ ! -e "@TERMUX_PREFIX@/bin/qemu-${DISTRO_ARCH/i686/i386}" ]; then
				msg
				msg "${BRED}Error: package 'qemu-user-${DISTRO_ARCH/i686/i386}' is not installed.${RST}"
				msg
				return 1
			fi

			qemu_arg="-q @TERMUX_PREFIX@/bin/qemu-${DISTRO_ARCH/i686/i386}"
		fi
	fi

	proot \
		$qemu_arg -L \
		--kernel-release=5.4.0-faked \
		--link2symlink \
		--kill-on-exit \
		--rootfs="${INSTALLED_ROOTFS_DIR}/${distro_name}" \
		--root-id \
		--cwd=/root \
		--bind=/dev \
		--bind="/dev/urandom:/dev/random" \
		--bind=/proc \
		--bind="/proc/self/fd:/dev/fd" \
		--bind="/proc/self/fd/0:/dev/stdin" \
		--bind="/proc/self/fd/1:/dev/stdout" \
		--bind="/proc/self/fd/2:/dev/stderr" \
		--bind=/sys \
		--bind="${INSTALLED_ROOTFS_DIR}/${distro_name}/proc/.loadavg:/proc/loadavg" \
		--bind="${INSTALLED_ROOTFS_DIR}/${distro_name}/proc/.stat:/proc/stat" \
		--bind="${INSTALLED_ROOTFS_DIR}/${distro_name}/proc/.uptime:/proc/uptime" \
		--bind="${INSTALLED_ROOTFS_DIR}/${distro_name}/proc/.version:/proc/version" \
		--bind="${INSTALLED_ROOTFS_DIR}/${distro_name}/proc/.vmstat:/proc/vmstat" \
		/usr/bin/env -i \
			"HOME=/root" \
			"LANG=C.UTF-8" \
			"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin" \
			"TERM=$TERM" \
			"TMPDIR=/tmp" \
			"$@"
}

# A function for preparing fake content for certain /proc
# entries which are known to be restricted on Android.
setup_fake_proc() {
	mkdir -p "${INSTALLED_ROOTFS_DIR}/${distro_name}/proc"
	chmod 700 "${INSTALLED_ROOTFS_DIR}/${distro_name}/proc"

	if [ ! -f "${INSTALLED_ROOTFS_DIR}/${distro_name}/proc/.loadavg" ]; then
		cat <<- EOF > "${INSTALLED_ROOTFS_DIR}/${distro_name}/proc/.loadavg"
		0.54 0.41 0.30 1/931 370386
		EOF
	fi

	if [ ! -f "${INSTALLED_ROOTFS_DIR}/${distro_name}/proc/.stat" ]; then
		cat <<- EOF > "${INSTALLED_ROOTFS_DIR}/${distro_name}/proc/.stat"
		cpu  1050008 127632 898432 43828767 37203 63 99244 0 0 0
		cpu0 212383 20476 204704 8389202 7253 42 12597 0 0 0
		cpu1 224452 24947 215570 8372502 8135 4 42768 0 0 0
		cpu2 222993 17440 200925 8424262 8069 9 17732 0 0 0
		cpu3 186835 8775 195974 8486330 5746 3 8360 0 0 0
		cpu4 107075 32886 48854 8688521 3995 4 5758 0 0 0
		cpu5 90733 20914 27798 1429573 2984 1 11419 0 0 0
		intr 53261351 0 686 1 0 0 1 12 31 1 20 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 7818 0 0 0 0 0 0 0 0 255 33 1912 33 0 0 0 0 0 0 3449534 2315885 2150546 2399277 696281 339300 22642 19371 0 0 0 0 0 0 0 0 0 0 0 2199 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 2445 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 162240 14293 2858 0 151709 151592 0 0 0 284534 0 0 0 0 0 0 0 0 0 0 0 0 0 0 185353 0 0 938962 0 0 0 0 736100 0 0 1 1209 27960 0 0 0 0 0 0 0 0 303 115968 452839 2 0 0 0 0 0 0 0 0 0 0 0 0 0 160361 8835 86413 1292 0 0 0 0 0 0 0 0 0 0 0 0 0 0 3592 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 6091 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 35667 0 0 156823 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 138 2667417 0 41 4008 952 16633 533480 0 0 0 0 0 0 262506 0 0 0 0 0 0 126 0 0 1558488 0 4 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 2 2 8 0 0 6 0 0 0 10 3 4 0 0 0 0 0 3 0 0 0 0 0 0 0 0 0 0 0 20 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 12 1 1 83806 0 1 1 0 1 0 1 1 319686 2 8 0 0 0 0 0 0 0 0 0 244534 0 1 10 9 0 10 112 107 40 221 0 0 0 144
		ctxt 90182396
		btime 1595203295
		processes 270853
		procs_running 2
		procs_blocked 0
		softirq 25293348 2883 7658936 40779 539155 497187 2864 1908702 7229194 279723 7133925
		EOF
	fi

	if [ ! -f "${INSTALLED_ROOTFS_DIR}/${distro_name}/proc/.uptime" ]; then
		cat <<- EOF > "${INSTALLED_ROOTFS_DIR}/${distro_name}/proc/.uptime"
		284684.56 513853.46
		EOF
	fi

	if [ ! -f "${INSTALLED_ROOTFS_DIR}/${distro_name}/proc/.version" ]; then
		cat <<- EOF > "${INSTALLED_ROOTFS_DIR}/${distro_name}/proc/.version"
		Linux version 5.4.0-faked (termux@androidos) (gcc version 4.9.x (Faked /proc/version by Proot-Distro) ) #1 SMP PREEMPT Fri Jul 10 00:00:00 UTC 2020
		EOF
	fi

	if [ ! -f "${INSTALLED_ROOTFS_DIR}/${distro_name}/proc/.vmstat" ]; then
		cat <<- EOF > "${INSTALLED_ROOTFS_DIR}/${distro_name}/proc/.vmstat"
		nr_free_pages 146031
		nr_zone_inactive_anon 196744
		nr_zone_active_anon 301503
		nr_zone_inactive_file 2457066
		nr_zone_active_file 729742
		nr_zone_unevictable 164
		nr_zone_write_pending 8
		nr_mlock 34
		nr_page_table_pages 6925
		nr_kernel_stack 13216
		nr_bounce 0
		nr_zspages 0
		nr_free_cma 0
		numa_hit 672391199
		numa_miss 0
		numa_foreign 0
		numa_interleave 62816
		numa_local 672391199
		numa_other 0
		nr_inactive_anon 196744
		nr_active_anon 301503
		nr_inactive_file 2457066
		nr_active_file 729742
		nr_unevictable 164
		nr_slab_reclaimable 132891
		nr_slab_unreclaimable 38582
		nr_isolated_anon 0
		nr_isolated_file 0
		workingset_nodes 25623
		workingset_refault 46689297
		workingset_activate 4043141
		workingset_restore 413848
		workingset_nodereclaim 35082
		nr_anon_pages 599893
		nr_mapped 136339
		nr_file_pages 3086333
		nr_dirty 8
		nr_writeback 0
		nr_writeback_temp 0
		nr_shmem 13743
		nr_shmem_hugepages 0
		nr_shmem_pmdmapped 0
		nr_file_hugepages 0
		nr_file_pmdmapped 0
		nr_anon_transparent_hugepages 57
		nr_unstable 0
		nr_vmscan_write 57250
		nr_vmscan_immediate_reclaim 2673
		nr_dirtied 79585373
		nr_written 72662315
		nr_kernel_misc_reclaimable 0
		nr_dirty_threshold 657954
		nr_dirty_background_threshold 328575
		pgpgin 372097889
		pgpgout 296950969
		pswpin 14675
		pswpout 59294
		pgalloc_dma 4
		pgalloc_dma32 101793210
		pgalloc_normal 614157703
		pgalloc_movable 0
		allocstall_dma 0
		allocstall_dma32 0
		allocstall_normal 184
		allocstall_movable 239
		pgskip_dma 0
		pgskip_dma32 0
		pgskip_normal 0
		pgskip_movable 0
		pgfree 716918803
		pgactivate 68768195
		pgdeactivate 7278211
		pglazyfree 1398441
		pgfault 491284262
		pgmajfault 86567
		pglazyfreed 1000581
		pgrefill 7551461
		pgsteal_kswapd 130545619
		pgsteal_direct 205772
		pgscan_kswapd 131219641
		pgscan_direct 207173
		pgscan_direct_throttle 0
		zone_reclaim_failed 0
		pginodesteal 8055
		slabs_scanned 9977903
		kswapd_inodesteal 13337022
		kswapd_low_wmark_hit_quickly 33796
		kswapd_high_wmark_hit_quickly 3948
		pageoutrun 43580
		pgrotated 200299
		drop_pagecache 0
		drop_slab 0
		oom_kill 0
		numa_pte_updates 0
		numa_huge_pte_updates 0
		numa_hint_faults 0
		numa_hint_faults_local 0
		numa_pages_migrated 0
		pgmigrate_success 768502
		pgmigrate_fail 1670
		compact_migrate_scanned 1288646
		compact_free_scanned 44388226
		compact_isolated 1575815
		compact_stall 863
		compact_fail 392
		compact_success 471
		compact_daemon_wake 975
		compact_daemon_migrate_scanned 613634
		compact_daemon_free_scanned 26884944
		htlb_buddy_alloc_success 0
		htlb_buddy_alloc_fail 0
		unevictable_pgs_culled 258910
		unevictable_pgs_scanned 3690
		unevictable_pgs_rescued 200643
		unevictable_pgs_mlocked 199204
		unevictable_pgs_munlocked 199164
		unevictable_pgs_cleared 6
		unevictable_pgs_stranded 6
		thp_fault_alloc 10655
		thp_fault_fallback 130
		thp_collapse_alloc 655
		thp_collapse_alloc_failed 50
		thp_file_alloc 0
		thp_file_mapped 0
		thp_split_page 612
		thp_split_page_failed 0
		thp_deferred_split_page 11238
		thp_split_pmd 632
		thp_split_pud 0
		thp_zero_page_alloc 2
		thp_zero_page_alloc_failed 0
		thp_swpout 4
		thp_swpout_fallback 0
		balloon_inflate 0
		balloon_deflate 0
		balloon_migrate 0
		swap_ra 9661
		swap_ra_hit 7872
		EOF
	fi
}

# Usage info for command_install.
command_install_help() {
	msg
	msg "${BYELLOW}Usage: ${BCYAN}$PROGRAM_NAME ${GREEN}install ${CYAN}[${GREEN}DISTRIBUTION ALIAS${CYAN}]${RST}"
	msg
	msg "${CYAN}This command will create a fresh installation of specified Linux${RST}"
	msg "${CYAN}distribution.${RST}"
	msg
	msg "${CYAN}Options:${RST}"
	msg
	msg "  ${GREEN}--help               ${CYAN}- Show this help information.${RST}"
	msg
	msg "  ${GREEN}--override-alias [new alias]   ${CYAN}- Set a custom alias for installed${RST}"
	msg "                                   ${CYAN}distribution.${RST}"
	msg
	msg "${CYAN}Selected distribution should be referenced by alias which can be${RST}"
	msg "${CYAN}obtained by this command: ${GREEN}$PROGRAM_NAME list${RST}"
	msg
	show_version
	msg
}

#############################################################################
#
# FUNCTION TO UNINSTALL SPECIFIED DISTRIBUTION
#
# Just deletes the rootfs of the given distribution.
#
# Accepted agruments: $1 - name of distribution.
#
command_remove() {
	local distro_name

	if [ $# -ge 1 ]; then
		case "$1" in
			-h|--help)
				command_remove_help
				return 0
				;;
			*) distro_name="$1";;
		esac
	else
		msg
		msg "${BRED}Error: distribution alias is not specified.${RST}"
		command_remove_help
		return 1
	fi

	if [ -z "${SUPPORTED_DISTRIBUTIONS["$distro_name"]+x}" ]; then
		msg
		msg "${BRED}Error: unknown distribution '${YELLOW}${distro_name}${BRED}' was requested to be removed.${RST}"
		msg
		msg "${CYAN}Use '${GREEN}${PROGRAM_NAME} list${CYAN}' to see which distributions are supported.${RST}"
		msg
		return 1
	fi

	# Not using is_distro_installed() here as we only need to know
	# whether rootfs directory is available.
	if [ ! -d "${INSTALLED_ROOTFS_DIR}/${distro_name}" ]; then
		msg
		msg "${BRED}Error: distribution '${YELLOW}${distro_name}${BRED}' is not installed.${RST}"
		msg
		return 1
	fi

	# Delete plugin with overridden alias.
	if [ "${CMD_REMOVE_REQUESTED_RESET-false}" = "false" ] && [ -e "${DISTRO_PLUGINS_DIR}/${distro_name}.override.sh" ]; then
		msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Deleting ${DISTRO_PLUGINS_DIR}/${distro_name}.override.sh...${RST}"
		rm -f "${DISTRO_PLUGINS_DIR}/${distro_name}.override.sh"
	fi

	msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Wiping the rootfs of ${YELLOW}${SUPPORTED_DISTRIBUTIONS["$distro_name"]}${CYAN}...${RST}"
	# Attempt to restore permissions so directory can be removed without issues.
	chmod u+rwx -R "${INSTALLED_ROOTFS_DIR}/${distro_name}" > /dev/null 2>&1 || true
	# There is still chance for failure.
	if ! rm -rf "${INSTALLED_ROOTFS_DIR:?}/${distro_name:?}"; then
		msg "${BLUE}[${RED}!${BLUE}] ${CYAN}Finished with errors. Some files probably were not deleted.${RST}"
		return 1
	fi
}

# Usage info for command_remove.
command_remove_help() {
	msg
	msg "${BYELLOW}Usage: ${BCYAN}$PROGRAM_NAME ${GREEN}remove ${CYAN}[${GREEN}DISTRIBUTION ALIAS${CYAN}]${RST}"
	msg
	msg "${CYAN}This command will uninstall the specified Linux distribution.${RST}"
	msg
	msg "${CYAN}Be careful when using it because you will not be prompted for${RST}"
	msg "${CYAN}confirmation and all data saved within the distribution will${RST}"
	msg "${CYAN}instantly gone.${RST}"
	msg
	msg "${CYAN}Selected distribution should be referenced by alias which can be${RST}"
	msg "${CYAN}obtained by this command: ${GREEN}$PROGRAM_NAME list${RST}"
	msg
	show_version
	msg
}

#############################################################################
#
# FUNCTION TO REINSTALL THE GIVEN DISTRIBUTION
#
# Just a shortcut for command_remove && command_install.
#
# Accepted arguments: $1 - distribution name.
#
command_reset() {
	local distro_name

	if [ $# -ge 1 ]; then
		case "$1" in
			-h|--help)
				command_reset_help
				return 0
				;;
			*) distro_name="$1";;
		esac
	else
		msg
		msg "${BRED}Error: distribution alias is not specified.${RST}"
		command_reset_help
		return 1
	fi

	if [ -z "${SUPPORTED_DISTRIBUTIONS["$distro_name"]+x}" ]; then
		msg
		msg "${BRED}Error: unknown distribution '${YELLOW}${distro_name}${BRED}' was requested to be reinstalled.${RST}"
		msg
		msg "${CYAN}Use '${GREEN}${PROGRAM_NAME} list${CYAN}' to see which distributions are supported.${RST}"
		msg
		return 1
	fi

	if [ ! -d "${INSTALLED_ROOTFS_DIR}/${distro_name}" ]; then
		msg
		msg "${BRED}Error: distribution '${YELLOW}${distro_name}${BRED}' is not installed.${RST}"
		msg
		return 1
	fi

	CMD_REMOVE_REQUESTED_RESET="true" command_remove "$distro_name"
	command_install "$distro_name"
}

# Usage info for command_reset.
command_reset_help() {
	msg
	msg "${BYELLOW}Usage: ${BCYAN}$PROGRAM_NAME ${GREEN}reset ${CYAN}[${GREEN}DISTRIBUTION ALIAS${CYAN}]${RST}"
	msg
	msg "${CYAN}Reinstall the specified Linux distribution.${RST}"
	msg
	msg "${CYAN}Be careful when using it because you will not be prompted for${RST}"
	msg "${CYAN}confirmation and all data saved within the distribution will${RST}"
	msg "${CYAN}instantly gone.${RST}"
	msg
	msg "${CYAN}Selected distribution should be referenced by alias which can be${RST}"
	msg "${CYAN}obtained by this command: ${GREEN}$PROGRAM_NAME list${RST}"
	msg
	show_version
	msg
}

#############################################################################
#
# FUNCTION TO START SHELL OR EXECUTE COMMAND
#
# Starts root shell inside the rootfs of specified Linux distribution.
# If '--' with further arguments was specified, execute the root shell
# command and exit.
#
# Accepts arbitrary amount of arguments.
#
command_login() {
	local isolated_environment=false
	local use_termux_home=false
	local no_link2symlink=false
	local no_sysvipc=false
	local no_kill_on_exit=false
	local fix_low_ports=false
	local make_host_tmp_shared=false
	local distro_name=""
	local login_user="root"
	local -a custom_fs_bindings
	local need_qemu=false

	while (($# >= 1)); do
		case "$1" in
			--)
				shift 1
				break
				;;
			--help)
				command_login_help
				return 0
				;;
			--fix-low-ports)
				fix_low_ports=true
				;;
			--isolated)
				isolated_environment=true
				;;
			--termux-home)
				use_termux_home=true
				;;
			--shared-tmp)
				make_host_tmp_shared=true
				;;
			--bind)
				if [ $# -ge 2 ]; then
					shift 1

					if [ -z "$1" ]; then
						msg
						msg "${BRED}Error: argument to option '${YELLOW}--bind${BRED}' should not be empty.${RST}"
						command_login_help
						return 1
					fi

					custom_fs_bindings+=("$1")
				else
					msg
					msg "${BRED}Error: option '${YELLOW}$1${BRED}' requires an argument.${RST}"
					command_login_help
					return 1
				fi
				;;
			--no-link2symlink)
				no_link2symlink=true
				;;
			--no-sysvipc)
				no_sysvipc=true
				;;
			--no-kill-on-exit)
				no_kill_on_exit=true
				;;
			--user)
				if [ $# -ge 2 ]; then
					shift 1

					if [ -z "$1" ]; then
						msg
						msg "${BRED}Error: argument to option '${YELLOW}--user${BRED}' should not be empty.${RST}"
						command_login_help
						return 1
					fi

					login_user="$1"
				else
					msg
					msg "${BRED}Error: option '${YELLOW}$1${BRED}' requires an argument.${RST}"
					command_login_help
					return 1
				fi
				;;
			-*)
				msg
				msg "${BRED}Error: unknown option '${YELLOW}${1}${BRED}'.${RST}"
				command_login_help
				return 1
				;;
			*)
				if [ -z "$1" ]; then
					msg
					msg "${BRED}Error: you should not pass empty command line arguments.${RST}"
					command_login_help
					return 1
				fi

				if [ -z "$distro_name" ]; then
					distro_name="$1"
				else
					msg
					msg "${BRED}Error: unknown option '${YELLOW}${1}${BRED}'.${RST}"
					msg
					msg "${BRED}Error: you have already set distribution as '${YELLOW}${distro_name}${BRED}'.${RST}"
					command_login_help
					return 1
				fi
				;;
		esac
		shift 1
	done

	if [ -z "$distro_name" ]; then
		msg
		msg "${BRED}Error: you should at least specify a distribution in order to log in.${RST}"
		command_login_help
		return 1
	fi

	if is_distro_installed "$distro_name"; then
		if [ -d "${INSTALLED_ROOTFS_DIR}/${distro_name}/.l2s" ]; then
			export PROOT_L2S_DIR="${INSTALLED_ROOTFS_DIR}/${distro_name}/.l2s"
		fi
		unset LD_PRELOAD
		#export PROOT_NO_SECCOMP=1

		if [ $# -ge 1 ]; then
			# Wrap in quotes each argument to prevent word splitting.
			local -a shell_command_args
			for i in "$@"; do
				shell_command_args+=("'$i'")
			done

			set -- "/bin/su" "-l" "$login_user" "-c" "${shell_command_args[*]}"
		else
			set -- "/bin/su" "-l" "$login_user"
		fi

		# Setup the default environment as well as copy some variables
		# defined by Termux. Note that when copying variables, we don't
		# care whether they actually defined in Termux or not. If they
		# will be empty, this should not cause any issues.
		set -- "/usr/bin/env" "-i" \
			"HOME=/root" \
			"LANG=C.UTF-8" \
			"TERM=${TERM-xterm-256color}" \
			"$@"

		set -- "--rootfs=${INSTALLED_ROOTFS_DIR}/${distro_name}" "$@"

		# Setup QEMU when CPU architecture do not match the one of device.
		local target_arch
		if [ -f "${DISTRO_PLUGINS_DIR}/${distro_name}.sh" ]; then
			target_arch=$(. "${DISTRO_PLUGINS_DIR}/${distro_name}.sh"; echo "${DISTRO_ARCH}")
		elif [ -f "${DISTRO_PLUGINS_DIR}/${distro_name}.override.sh" ]; then
			target_arch=$(. "${DISTRO_PLUGINS_DIR}/${distro_name}.override.sh"; echo "${DISTRO_ARCH}")
		else
			# This should never happen.
			msg
			msg "${BRED}Error: missing plugin for distribution '${YELLOW}${distro_name}${BRED}'.${RST}"
			msg
			return 1
		fi

		if [ "$DEVICE_CPU_ARCH" != "$target_arch" ]; then
			need_qemu=true
			if ! [[ "$target_arch" =~ ^(aarch64|arm|i686|x86_64)$ ]]; then
				msg
				msg "${BRED}Error: DISTRO_ARCH has unknown value '$target_arch'. Valid values are: aarch64, arm, i686, x86_64."
				msg
				return 1
			fi

			# If CPU and host OS are 64bit, we can run 32bit guest OS without emulation.
			# Everything else requires emulator (QEMU).
			if ! [ "$DEVICE_CPU_ARCH" = "aarch64" ] && [ "$DISTRO_ARCH" = "arm" ] || \
				! [ "$DEVICE_CPU_ARCH" = "x86_64" ] && [ "$DISTRO_ARCH" = "i686" ]; then

				if [ ! -e "@TERMUX_PREFIX@/bin/qemu-${target_arch/i686/i386}" ]; then
					msg
					msg "${BRED}Error: package 'qemu-user-${target_arch/i686/i386}' is not installed.${RST}"
					msg
					return 1
				fi

				set -- "-q" "@TERMUX_PREFIX@/bin/qemu-${target_arch/i686/i386}" "$@"
			fi
		fi

		if ! $no_kill_on_exit; then
			# This option terminates all background processes on exit, so
			# proot can terminate freely.
			set -- "--kill-on-exit" "$@"
		else
			msg "${BRED}Warning: option '--no-kill-on-exit' is enabled. When exiting, your session will be blocked until all processes are terminated.${RST}"
		fi

		if ! $no_link2symlink; then
			# Support hardlinks.
			set -- "--link2symlink" "$@"
		fi

		if ! $no_sysvipc; then
			# Support System V IPC.
			set -- "--sysvipc" "$@"
		fi

		# Some devices have old kernels and GNU libc refuses to work on them.
		# Fix this behavior by reporting a fake up-to-date kernel version.
		set -- "--kernel-release=5.4.0-faked" "$@"

		# Fix lstat to prevent dpkg symlink size warnings
		set -- "-L" "$@"

		# Simulate root so we can switch users.
		set -- "--cwd=/root" "$@"
		set -- "--root-id" "$@"

		# Core file systems that should always be present.
		set -- "--bind=/dev" "$@"
		set -- "--bind=/dev/urandom:/dev/random" "$@"
		set -- "--bind=/proc" "$@"
		set -- "--bind=/proc/self/fd:/dev/fd" "$@"
		set -- "--bind=/proc/self/fd/0:/dev/stdin" "$@"
		set -- "--bind=/proc/self/fd/1:/dev/stdout" "$@"
		set -- "--bind=/proc/self/fd/2:/dev/stderr" "$@"
		set -- "--bind=/sys" "$@"

		# Ensure that we can bind fake /proc entries.
		setup_fake_proc

		# Fake /proc/loadavg if necessary.
		if ! cat /proc/loadavg > /dev/null 2>&1; then
			set -- "--bind=${INSTALLED_ROOTFS_DIR}/${distro_name}/proc/.loadavg:/proc/loadavg" "$@"
		fi

		# Fake /proc/stat if necessary.
		if ! cat /proc/stat > /dev/null 2>&1; then
			set -- "--bind=${INSTALLED_ROOTFS_DIR}/${distro_name}/proc/.stat:/proc/stat" "$@"
		fi

		# Fake /proc/uptime if necessary.
		if ! cat /proc/uptime > /dev/null 2>&1; then
			set -- "--bind=${INSTALLED_ROOTFS_DIR}/${distro_name}/proc/.uptime:/proc/uptime" "$@"
		fi

		# Fake /proc/version if necessary.
		if ! cat /proc/version > /dev/null 2>&1; then
			set -- "--bind=${INSTALLED_ROOTFS_DIR}/${distro_name}/proc/.version:/proc/version" "$@"
		fi

		# Fake /proc/vmstat if necessary.
		if ! cat /proc/vmstat > /dev/null 2>&1; then
			set -- "--bind=${INSTALLED_ROOTFS_DIR}/${distro_name}/proc/.vmstat:/proc/vmstat" "$@"
		fi

		# Bind /tmp to /dev/shm.
		if [ ! -d "${INSTALLED_ROOTFS_DIR}/${distro_name}/tmp" ]; then
			mkdir -p "${INSTALLED_ROOTFS_DIR}/${distro_name}/tmp"
		fi
		set -- "--bind=${INSTALLED_ROOTFS_DIR}/${distro_name}/tmp:/dev/shm" "$@"

		# When running in non-isolated mode, provide some bindings specific
		# to Android and Termux so user can interact with host file system.
		if ! $isolated_environment; then
			set -- "--bind=/data/dalvik-cache" "$@"
			set -- "--bind=/data/data/com.termux/cache" "$@"
			set -- "--bind=/data/data/com.termux/files/home" "$@"
			set -- "--bind=/storage" "$@"
			set -- "--bind=/storage/self/primary:/sdcard" "$@"
		fi

		# When using QEMU, we need some host files even in isolated mode.
		if ! $isolated_environment || $need_qemu; then
			if [ -d "/apex" ]; then
				set -- "--bind=/apex" "$@"
			fi
			if [ -e "/linkerconfig/ld.config.txt" ]; then
				set -- "--bind=/linkerconfig/ld.config.txt" "$@"
			fi
			set -- "--bind=/data/data/com.termux/files/usr" "$@"
			set -- "--bind=/system" "$@"
			set -- "--bind=/vendor" "$@"
			if [ -f "/plat_property_contexts" ]; then
				set -- "--bind=/plat_property_contexts" "$@"
			fi
			if [ -f "/property_contexts" ]; then
				set -- "--bind=/property_contexts" "$@"
			fi
		fi

		# Use Termux home directory if requested.
		# Ignores --isolated.
		if $use_termux_home; then
			if [ "$login_user" = "root" ]; then
				set -- "--bind=@TERMUX_HOME@:/root" "$@"
			else
				if [ -f "${INSTALLED_ROOTFS_DIR}/${distro_name}/etc/passwd" ]; then
					local user_home
					user_home=$(grep -P "^${login_user}:" "${INSTALLED_ROOTFS_DIR}/${distro_name}/etc/passwd" | cut -d: -f 6)

					if [ -z "$user_home" ]; then
						user_home="/home/${login_user}"
					fi

					set -- "--bind=@TERMUX_HOME@:${user_home}" "$@"
				else
					set -- "--bind=@TERMUX_HOME@:/home/${login_user}" "$@"
				fi
			fi
		fi

		# Bind the tmp folder from the host system to the guest system
		# Ignores --isolated.
		if $make_host_tmp_shared; then
			set -- "--bind=@TERMUX_PREFIX@/tmp:/tmp" "$@"
		fi

		# Bind custom file systems.
		local bnd
		for bnd in "${custom_fs_bindings[@]}"; do
			set -- "--bind=${bnd}" "$@"
		done

		# Modify bindings to protected ports to use a higher port number.
		if $fix_low_ports; then
			set -- "-p" "$@"
		fi

		exec proot "$@"
	else
		if [ -z "${SUPPORTED_DISTRIBUTIONS["$distro_name"]+x}" ]; then
			msg
			msg "${BRED}Error: cannot log in into unknown distribution '${YELLOW}${distro_name}${BRED}'.${RST}"
			msg
			msg "${CYAN}Use '${GREEN}${PROGRAM_NAME} list${CYAN}' to see which distributions are supported.${RST}"
			msg
		else
			msg
			msg "${BRED}Error: distribution '${YELLOW}${distro_name}${BRED}' is not installed.${RST}"
			msg
			msg "${CYAN}Install it with: ${GREEN}${PROGRAM_NAME} install ${distro_name}${RST}"
			msg
		fi
		return 1
	fi
}

# Usage info for command_login.
command_login_help() {
	msg
	msg "${BYELLOW}Usage: ${BCYAN}$PROGRAM_NAME ${GREEN}login ${CYAN}[${GREEN}OPTIONS${CYAN}] [${GREEN}DISTRO ALIAS${CYAN}] [${GREEN}--${CYAN}[${GREEN}COMMAND${CYAN}]]${RST}"
	msg
	msg "${CYAN}This command will launch a login shell for the specified${RST}"
	msg "${CYAN}distribution if no additional arguments were given, otherwise${RST}"
	msg "${CYAN}it will execute the given command and exit.${RST}"
	msg
	msg "${CYAN}Options:${RST}"
	msg
	msg "  ${GREEN}--help               ${CYAN}- Show this help information.${RST}"
	msg
	msg "  ${GREEN}--user [user]        ${CYAN}- Login as specified user instead of 'root'.${RST}"
	msg
	msg "  ${GREEN}--fix-low-ports      ${CYAN}- Modify bindings to protected ports to use${RST}"
	msg "                         ${CYAN}a higher port number.${RST}"
	msg
	msg "  ${GREEN}--isolated           ${CYAN}- Run isolated environment without access${RST}"
	msg "                         ${CYAN}to host file system.${RST}"
	msg
	msg "  ${GREEN}--termux-home        ${CYAN}- Mount Termux home directory to /root.${RST}"
	msg "                         ${CYAN}Takes priority over '${GREEN}--isolated${CYAN}' option.${RST}"
	msg
	msg "  ${GREEN}--shared-tmp         ${CYAN}- Mount Termux temp directory to /tmp.${RST}"
	msg "                         ${CYAN}Takes priority over '${GREEN}--isolated${CYAN}' option.${RST}"
	msg
	msg "  ${GREEN}--bind [path:path]   ${CYAN}- Custom file system binding. Can be specified${RST}"
	msg "                         ${CYAN}multiple times.${RST}"
	msg "                         ${CYAN}Takes priority over '${GREEN}--isolated${CYAN}' option.${RST}"
	msg
	msg "  ${GREEN}--no-link2symlink    ${CYAN}- Disable hardlink emulation by proot.${RST}"
	msg "                         ${CYAN}Adviseable only on devices with SELinux${RST}"
	msg "                         ${CYAN}in permissive mode.${RST}"
	msg
	msg "  ${GREEN}--no-sysvipc         ${CYAN}- Disable System V IPC emulation by proot.${RST}"
	msg
	msg "  ${GREEN}--no-kill-on-exit    ${CYAN}- Wait until all running processes will finish${RST}"
	msg "                         ${CYAN}before exiting. This will cause proot to${RST}"
	msg "                         ${CYAN}freeze if you are running daemons.${RST}"
	msg
	msg "${CYAN}Put '${GREEN}--${CYAN}' if you wish to stop command line processing and pass${RST}"
	msg "${CYAN}options as shell arguments.${RST}"
	msg
	msg "${CYAN}Selected distribution should be referenced by alias which can be${RST}"
	msg "${CYAN}obtained by this command: ${GREEN}$PROGRAM_NAME list${RST}"
	msg
	msg "${CYAN}If no '${GREEN}--isolated${CYAN}' option given, the following host directories${RST}"
	msg "${CYAN}will be available:${RST}"
	msg
	msg "  ${CYAN}* ${YELLOW}/apex ${CYAN}(only Android 10+)${RST}"
	msg "  ${CYAN}* ${YELLOW}/data/dalvik-cache${RST}"
	msg "  ${CYAN}* ${YELLOW}/data/data/com.termux${RST}"
	msg "  ${CYAN}* ${YELLOW}/sdcard${RST}"
	msg "  ${CYAN}* ${YELLOW}/storage${RST}"
	msg "  ${CYAN}* ${YELLOW}/system${RST}"
	msg "  ${CYAN}* ${YELLOW}/vendor${RST}"
	msg
	msg "${CYAN}This should be enough to get Termux utilities like termux-api or${RST}"
	msg "${CYAN}termux-open get working. If they do not work for some reason,${RST}"
	msg "${CYAN}check if these files are sourced by your shell:${RST}"
	msg
	msg "  ${CYAN}* ${YELLOW}/etc/profile.d/termux-proot.sh${RST}"
	msg "  ${CYAN}* ${YELLOW}/etc/profile${RST}"
	msg
	msg "${CYAN}Also check whether they define variables like ANDROID_DATA,${RST}"
	msg "${CYAN}ANDROID_ROOT, BOOTCLASSPATH and others which are usually set${RST}"
	msg "${CYAN}in Termux sessions.${RST}"
	msg
	show_version
	msg
}

#############################################################################
#
# FUNCTION TO LIST THE SUPPORTED DISTRIBUTIONS
#
# Shows the list of distributions which this utility can handle. Also print
# their installation status.
#
command_list() {
	msg
	if [ -z "${!SUPPORTED_DISTRIBUTIONS[*]}" ]; then
		msg "${YELLOW}You do not have any distribution plugins configured.${RST}"
		msg
		msg "${YELLOW}Please check the directory '$DISTRO_PLUGINS_DIR'.${RST}"
	else
		msg "${CYAN}Supported distributions:${RST}"

		local i
		for i in $(echo "${!SUPPORTED_DISTRIBUTIONS[@]}" | tr ' ' '\n' | sort -d); do
			msg
			msg "  ${CYAN}* ${YELLOW}${SUPPORTED_DISTRIBUTIONS[$i]}${RST}"
			msg
			msg "    ${CYAN}Alias: ${YELLOW}${i}${RST}"
			if is_distro_installed "$i"; then
				msg "    ${CYAN}Status: ${GREEN}installed${RST}"
			else
				msg "    ${CYAN}Status: ${RED}NOT installed${RST}"
			fi
			if [ -n "${SUPPORTED_DISTRIBUTIONS_COMMENTS["${i}"]+x}" ]; then
				msg "    ${CYAN}Comment: ${SUPPORTED_DISTRIBUTIONS_COMMENTS["${i}"]}${RST}"
			fi
		done

		msg
		msg "${CYAN}Install selected one with: ${GREEN}${PROGRAM_NAME} install <alias>${RST}"
	fi
	msg
}

#############################################################################
#
# FUNCTION TO BACKUP A SPECIFIED DISTRIBUTION
#
# Backup a specified distribution installation by making a tarball.
#
command_backup() {
	local distro_name=""
	local tarball_file_path=""

	while (($# >= 1)); do
		case "$1" in
			--)
				shift 1
				break
				;;
			--help)
				command_backup_help
				return 0
				;;
			--output)
				if [ $# -ge 2 ]; then
					shift 1

					if [ -z "$1" ]; then
						msg
						msg "${BRED}Error: argument to option '${YELLOW}--output${BRED}' should not be empty.${RST}"
						command_backup_help
						return 1
					fi

					tarball_file_path="$1"
				else
					msg
					msg "${BRED}Error: option '${YELLOW}$1${BRED}' requires an argument.${RST}"
					command_backup_help
					return 1
				fi
				;;
			-*)
				msg
				msg "${BRED}Error: unknown option '${YELLOW}${1}${BRED}'.${RST}"
				command_backup_help
				return 1
				;;
			*)
				if [ -z "$1" ]; then
					msg
					msg "${BRED}Error: you should not pass empty command line arguments.${RST}"
					command_backup_help
					return 1
				fi

				if [ -z "$distro_name" ]; then
					distro_name="$1"
				else
					msg
					msg "${BRED}Error: unknown option '${YELLOW}${1}${BRED}'.${RST}"
					msg
					msg "${BRED}Error: you have already set distribution as '${YELLOW}${distro_name}${BRED}'.${RST}"
					command_backup_help
					return 1
				fi
				;;
		esac
		shift 1
	done

	if [ -z "$distro_name" ]; then
		msg
		msg "${BRED}Error: you should specify a distribution which you want to back up.${RST}"
		command_backup_help
		return 1
	fi

	if [ -z "${SUPPORTED_DISTRIBUTIONS["$distro_name"]+x}" ]; then
		msg
		msg "${BRED}Error: unknown distribution '${YELLOW}${distro_name}${BRED}' was requested for backing up.${RST}"
		msg
		msg "${CYAN}Use '${GREEN}${PROGRAM_NAME} list${CYAN}' to see which distributions are supported.${RST}"
		msg
		return 1
	fi

	if ! is_distro_installed "$distro_name"; then
		msg
		msg "${BRED}Error: distribution '${YELLOW}${distro_name}${BRED}' is not installed.${RST}"
		msg
		return 1
	fi

	msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Backing up ${YELLOW}${SUPPORTED_DISTRIBUTIONS["$distro_name"]}${CYAN}...${RST}"

	if [ -z "$tarball_file_path" ]; then
		msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Tarball will be written to stdout.${RST}"

		if [ -t 1 ]; then
			msg
			msg "${BRED}Error: tarball cannot be printed to console, please use option '${YELLOW}--output${BRED}' or pipe it to another program.${RST}"
			msg
			return 1
		fi
	else
		msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Tarball will be written to '${tarball_file_path}'.${RST}"

		if [ -d "$tarball_file_path" ]; then
			msg
			msg "${BRED}Error: cannot write to '${YELLOW}${tarball_file_path}${YELLOW}' - path is a directory.${RST}"
			command_backup_help
			return 1
		fi

		if [ -f "$tarball_file_path" ]; then
			msg
			msg "${BRED}Error: file '${YELLOW}${tarball_file_path}${YELLOW}' already exist, please specify a different name.${RST}"
			command_backup_help
			return 1
		fi
	fi

	msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Fixing file permissions in rootfs...${RST}"
	# Ensure we can read all files.
	find "${INSTALLED_ROOTFS_DIR}/${distro_name}" -type d -print0 | xargs -0 -r chmod u+rx
	find "${INSTALLED_ROOTFS_DIR}/${distro_name}" -type f -executable -print0 | xargs -0 -r chmod u+rx
	find "${INSTALLED_ROOTFS_DIR}/${distro_name}" -type f ! -executable -print0 | xargs -0 -r chmod u+r

	local distro_plugin_script="${distro_name}.sh"
	if [ ! -f "${DISTRO_PLUGINS_DIR}/${distro_plugin_script}" ]; then
		# Alt name.
		distro_plugin_script="${distro_name}.override.sh"
	fi

	msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Archiving rootfs...${RST}"
	if [ -n "$tarball_file_path" ]; then
		tar -c --auto-compress \
			--warning=no-file-ignored \
			-f "$tarball_file_path" \
			-C "${DISTRO_PLUGINS_DIR}/../" "$(basename "$DISTRO_PLUGINS_DIR")/${distro_plugin_script}" \
			-C "${INSTALLED_ROOTFS_DIR}/../" "$(basename "$INSTALLED_ROOTFS_DIR")/${distro_name}"
	else
		tar -c \
			--warning=no-file-ignored \
			-C "${DISTRO_PLUGINS_DIR}/../" "$(basename "$DISTRO_PLUGINS_DIR")/${distro_plugin_script}" \
			-C "${INSTALLED_ROOTFS_DIR}/../" "$(basename "$INSTALLED_ROOTFS_DIR")/${distro_name}"
	fi
	msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Finished successfully.${RST}"
}

# Usage info for command_backup.
command_backup_help() {
	msg
	msg "${BYELLOW}Usage: ${BCYAN}$PROGRAM_NAME ${GREEN}backup ${CYAN}[${GREEN}DISTRIBUTION ALIAS${CYAN}]${RST}"
	msg
	msg "${CYAN}Back up a specified distribution installation into tarball.${RST}"
	msg
	msg "${CYAN}Options:${RST}"
	msg
	msg "  ${GREEN}--help               ${CYAN}- Show this help information.${RST}"
	msg
	msg "  ${GREEN}--output [path]      ${CYAN}- Write tarball to specified file.${RST}"
	msg "                         ${CYAN}If not specified, the tarball will be${RST}"
	msg "                         ${CYAN}printed to stdout. File extension affects${RST}"
	msg "                         ${CYAN}used compression (e.g. gz, bz2, xz)."
	msg "                         ${CYAN}Backup sent to stdout is not compressed.${RST}"
	msg
	msg "${CYAN}Selected distribution should be referenced by alias which can be${RST}"
	msg "${CYAN}obtained by this command: ${GREEN}$PROGRAM_NAME list${RST}"
	msg
	show_version
	msg
}

#############################################################################
#
# FUNCTION TO RESTORE A SPECIFIED DISTRIBUTION
#
# Restore a specified distribution installation from the backup (tarball).
#
command_restore() {
	local tarball_file_path=""

	if [ $# -ge 1 ]; then
		case "$1" in
			-h|--help)
				command_restore_help
				return 0
				;;
			*) tarball_file_path="$1";;
		esac
	else
		if [ -t 0 ]; then
			msg
			msg "${BRED}Error: tarball path is not specified and it looks like nothing is being piped via stdin.${RST}"
			command_restore_help
			return 1
		fi
	fi

	if [ -n "$tarball_file_path" ]; then
		if [ ! -e "$tarball_file_path" ]; then
			msg
			msg "${BRED}Error: file '${YELLOW}${tarball_file_path}${YELLOW}' is not found.${RST}"
			command_restore_help
			return 1
		fi

		if [ -d "$tarball_file_path" ]; then
			msg
			msg "${BRED}Error: path '${YELLOW}${tarball_file_path}${YELLOW}' is a directory.${RST}"
			command_restore_help
			return 1
		fi
	fi

	local success
	msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Extracting tarball...${RST}"
	if [ -n "$tarball_file_path" ]; then
		if ! tar -x --auto-compress -f "$tarball_file_path" \
			--recursive-unlink --preserve-permissions \
			-C "${DISTRO_PLUGINS_DIR}/../" "$(basename "${DISTRO_PLUGINS_DIR}")/" \
			-C "${INSTALLED_ROOTFS_DIR}/../" "$(basename "${INSTALLED_ROOTFS_DIR}")/"; then
			success=false
		else
			success=true
		fi
	else
		if ! tar -x --recursive-unlink --preserve-permissions \
			-C "${DISTRO_PLUGINS_DIR}/../" "$(basename "${DISTRO_PLUGINS_DIR}")/" \
			-C "${INSTALLED_ROOTFS_DIR}/../" "$(basename "${INSTALLED_ROOTFS_DIR}")/"; then
			success=false
		else
			success=true
		fi
	fi

	if $success; then
		msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Finished...${RST}"
	else
		msg "${BLUE}[${RED}!${BLUE}] ${CYAN}Failure.${RST}"
		msg
		msg "${BRED}Failed to restore distribution from the given tarball.${RST}"
		msg
		msg "${BRED}Possibly that tarball is corrupted or was not made by Proot-Distro.${RST}"
		msg
	fi
}

# Usage info for command_restore.
command_restore_help() {
	msg
	msg "${BYELLOW}Usage: ${BCYAN}$PROGRAM_NAME ${GREEN}restore ${CYAN}[${GREEN}FILENAME.TAR${CYAN}]${RST}"
	msg
	msg "${CYAN}Restore distribution installation from a specified tarball. If${RST}"
	msg "${CYAN}file name is not specified, it will be assumed that tarball is${RST}"
	msg "${CYAN}being piped from stdin.${RST}"
	msg
	msg "${CYAN}Archive compression is determined automatically from the file${RST}"
	msg "${CYAN}extension. If archive content is piped, it is expected that${RST}"
	msg "${CYAN}data is not compressed.${RST}"
	msg
	msg "${CYAN}Important note: there are no any sanity check being performed${RST}"
	msg "${CYAN}on the supplied tarballs. Be careful when using this command as${RST}"
	msg "${CYAN}data loss may happen when the wrong tarball has been used.${RST}"
	msg
	show_version
	msg
}

#############################################################################
#
# FUNCTION TO CLEAR DLCACHE
#
# Removes all cached downloads.
#
command_clear_cache() {
	if [ $# -ge 1 ]; then
		case "$1" in
			-h|--help)
				command_clear_cache_help
				return 0
				;;
			*)
				msg
				msg "${BRED}Error: unknown option '${YELLOW}${1}${BRED}'.${RST}"
				command_clear_cache_help
				return 1
				;;
		esac
	fi

	if ! ls -la "${DOWNLOAD_CACHE_DIR}"/* > /dev/null 2>&1; then
		msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Download cache is empty.${RST}"
	else
		local size_of_cache
		size_of_cache="$(du -d 0 -h -a ${DOWNLOAD_CACHE_DIR} | awk '{$2=$2};1' | cut -d " " -f 1)"

		msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Clearing cache files...${RST}"

		local filename
		while read -r filename; do
			msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Deleting ${CYAN}'${filename}'${RST}"
			rm -f "${filename}"
		done < <(find "${DOWNLOAD_CACHE_DIR}" -type f)

		msg "${BLUE}[${GREEN}*${BLUE}] ${CYAN}Reclaimed ${size_of_cache} of disk space.${RST}"
	fi
}

# Usage info for command_clear_cache.
command_clear_cache_help() {
	msg
	msg "${BYELLOW}Usage: ${BCYAN}$PROGRAM_NAME ${GREEN}clear-cache${RST}"
	msg
	msg "${CYAN}This command will reclaim some disk space by deleting cached${RST}"
	msg "${CYAN}distribution rootfs tarballs.${RST}"
	msg
	show_version
	msg
}

#############################################################################
#
# FUNCTION TO PRINT UTILITY USAGE INFORMATION
#
# Prints a basic overview of the available commands and list of supported
# distributions.
#
command_help() {
	msg
	msg "${BYELLOW}Usage: ${BCYAN}$PROGRAM_NAME${CYAN} [${GREEN}COMMAND${CYAN}] [${GREEN}ARGUMENTS${CYAN}]${RST}"
	msg
	msg "${CYAN}Utility to manage proot'ed Linux distributions inside Termux.${RST}"
	msg
	msg "${CYAN}List of the available commands:${RST}"
	msg
	msg "  ${GREEN}help         ${CYAN}- Show this help information.${RST}"
	msg
	msg "  ${GREEN}backup       ${CYAN}- Backup a specified distribution.${RST}"
	msg
	msg "  ${GREEN}install      ${CYAN}- Install a specified distribution.${RST}"
	msg
	msg "  ${GREEN}list         ${CYAN}- List supported distributions and their${RST}"
	msg "                 ${CYAN}installation status.${RST}"
	msg
	msg "  ${GREEN}login        ${CYAN}- Start login shell for the specified distribution.${RST}"
	msg
	msg "  ${GREEN}remove       ${CYAN}- Delete a specified distribution.${RST}"
	msg "                 ${RED}WARNING: this command destroys data!${RST}"
	msg
	msg "  ${GREEN}reset        ${CYAN}- Reinstall from scratch a specified distribution.${RST}"
	msg "                 ${RED}WARNING: this command destroys data!${RST}"
	msg
	msg "  ${GREEN}restore      ${CYAN}- Restore a specified distribution.${RST}"
	msg "                 ${RED}WARNING: this command destroys data!${RST}"
	msg
	msg "  ${GREEN}clear-cache  ${CYAN}- Clear cache of downloaded files. ${RST}"
	msg
	msg "${CYAN}Each of commands has its own help information. To view it, just${RST}"
	msg "${CYAN}supply a '${GREEN}--help${CYAN}' argument to chosen command.${RST}"
	msg
	msg "${CYAN}Hint: type command '${GREEN}${PROGRAM_NAME} list${CYAN}' to get a list of the${RST}"
	msg "${CYAN}supported distributions. Pick a distro alias and run the next${RST}"
	msg "${CYAN}command to install it: ${GREEN}${PROGRAM_NAME} install <alias>${RST}"
	msg
	msg "${CYAN}Runtime data is stored at this location:${RST}"
	msg "${CYAN}${RUNTIME_DIR}${RST}"
	msg
	msg "${CYAN}If you have issues with proot during installation or login, try${RST}"
	msg "${CYAN}to set '${GREEN}PROOT_NO_SECCOMP=1${CYAN}' environment variable.${RST}"
	msg
	show_version
	msg
}

#############################################################################
#
# FUNCTION TO PRINT VERSION STRING
#
# Prints version & author information. Used in functions for displaying
# usage info.
#
show_version() {
	msg "${ICYAN}Proot-Distro v${PROGRAM_VERSION} by @xeffyr.${RST}"
}

#############################################################################
#
# ENTRY POINT
#
# 1. Check for dependencies. Assume that package 'coreutils' is always
#    available.
# 2. Check all available distribution plug-ins.
# 3. Handle the requested commands or show help when '-h/--help/help' were
#    given. Further command line processing is offloaded to requested command.
#

# This will be executed when signal HUP/INT/TERM is received.
trap 'echo -e "\\r${BLUE}[${RED}!${BLUE}] ${CYAN}Exiting immediately as requested.${RST}"; exit 1;' HUP INT TERM

for i in awk bzip2 curl find gzip proot sed tar xz; do
	if [ -z "$(command -v "$i")" ]; then
		msg
		msg "${BRED}Utility '${i}' is not installed. Cannot continue.${RST}"
		msg
		exit 1
	fi
done
unset i

# Determine a CPU architecture of device.
case "$(uname -m)" in
	# Note: armv8l means that device is running 32bit OS on 64bit CPU.
	armv7l|armv8l) DEVICE_CPU_ARCH="arm";;
	*) DEVICE_CPU_ARCH=$(uname -m);;
esac
DISTRO_ARCH=$DEVICE_CPU_ARCH

# Verify architecture if possible - avoid running under linux32 or similar.
if [ -x "@TERMUX_PREFIX@/bin/dpkg" ]; then
	if [ "$DEVICE_CPU_ARCH" != "$("@TERMUX_PREFIX@"/bin/dpkg --print-architecture)" ]; then
		msg
		msg "${BRED}CPU architecture reported by system doesn't match architecture of Termux packages. Make sure you are not using 'linux32' or similar utilities.${RST}"
		msg
		exit 1
	fi
fi

declare -A TARBALL_URL TARBALL_SHA256
declare -A SUPPORTED_DISTRIBUTIONS
declare -A SUPPORTED_DISTRIBUTIONS_COMMENTS
while read -r filename; do
	distro_name=$(. "$filename"; echo "${DISTRO_NAME-}")
	distro_comment=$(. "$filename"; echo "${DISTRO_COMMENT-}")
	# May have 2 name formats:
	# * alias.override.sh
	# * alias.sh
	# but we need to treat both as 'alias'.
	distro_alias=${filename%%.override.sh}
	distro_alias=${distro_alias%%.sh}
	distro_alias=$(basename "$distro_alias")

	# We getting distribution name from $DISTRO_NAME which
	# should be set in plug-in.
	if [ -z "$distro_name" ]; then
		msg
		msg "${BRED}Error: no DISTRO_NAME defined in '${YELLOW}${filename}${BRED}'.${RST}"
		msg
		exit 1
	fi

	SUPPORTED_DISTRIBUTIONS["$distro_alias"]="$distro_name"
	[ -n "$distro_comment" ] && SUPPORTED_DISTRIBUTIONS_COMMENTS["$distro_alias"]="$distro_comment"
done < <(find "$DISTRO_PLUGINS_DIR" -maxdepth 1 -type f -iname "*.sh" 2>/dev/null)
unset distro_name distro_alias

if [ $# -ge 1 ]; then
	case "$1" in
		-h|--help|help) shift 1; command_help;;
		backup) shift 1; command_backup "$@";;
		install) shift 1; command_install "$@";;
		list) shift 1; command_list;;
		login) shift 1; command_login "$@";;
		remove) shift 1; CMD_REMOVE_REQUESTED_RESET="false" command_remove "$@";;
		clear-cache) shift 1; command_clear_cache "$@";;
		reset) shift 1; command_reset "$@";;
		restore) shift 1; command_restore "$@";;
		*)
			msg
			msg "${BRED}Error: unknown command '${YELLOW}$1${BRED}'.${RST}"
			msg
			msg "${CYAN}Run '${GREEN}${PROGRAM_NAME} help${CYAN}' to see the list of available commands.${RST}"
			msg
			exit 1
			;;
	esac
else
	msg
	msg "${BRED}Error: no command provided.${RST}"
	command_help
fi

exit 0
