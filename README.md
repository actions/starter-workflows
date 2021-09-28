# PRoot Distro

A Bash script wrapper for utility [proot] for easy management of chroot-based
Linux distribution installations. It does not require root or any special ROM,
kernel, etc. Everything you need to get started is the latest version of
[Termux] application. See [Installing](#installation) for details.

PRoot Distro is not a virtual machine, neither a traditional chroot. It shares
the same kernel as your Android system, so do not even try to update it through
package manager - this will not work.

This script should never be run as root user. If you do so, file permissions
and SELinux labels could get messed up. There also possibility of damaging
system if being executed as root. For safety, PRoot Distro checks the user id
before run and refuses to work if detected user id `0` (root).

***

## Supported distributions

PRoot Distro provides support only one version of distribution types, i.e. one
of stable, LTS or rolling-release. Support of versioned distributions ended
with branch 2.x. If you need a custom version, you will need to add it on your
own. See [Adding distribution](#adding-distribution).

Here are the supported distributions:

* Alpine Linux (edge)
* Arch Linux / Arch Linux 32 / Arch Linux ARM
* Debian (stable)
* Fedora 34
* Gentoo
* OpenSUSE (Tumbleweed)
* Ubuntu (21.04)
* Void Linux

All systems come in a bare-minumum variant, typically consisting of package
manager, shell, coreutils, util-linux and few more. Extended functionality
like shell completion or package install suggestions should be configured
manually.

If desired distribution is not in the list, you can request it.

## Installing

With package manager:
```
pkg install proot-distro
```

With git:
```
pkg install git
git clone https://github.com/termux/proot-distro
cd proot-distro
./install.sh
```

Dependencies: bash, bzip2, coreutils, curl, findutils, gzip, ncurses-utils,
proot, sed, tar, xz-utils

## Functionality overview

PRoot Distro aims to provide all-in-one functionality for managing the
installed distributions: installation, de-installation, backup, restore, login.
Each action is defined through command. Each command accepts its unique set
of options, specific to the task that it performs.

Usage basics:
```
proot-distro <command> <arguments>
```

Where `<command>` is a proot-distro action command (see below to learn what
is available) and `<arguments>` is a list of options specific to given command.

Example of installing the distribution:
```
proot-distro install debian
```

Known distributions are defined through plug-in scripts, which define URLs
from where root file system archive will be downloaded and set of checksums
for integrity check. Plug-ins also can define a set of commands which would
be executed during distribution installation.

See [Adding distribution](#adding-distribution) to learn more how to add own
distribution to PRoot Distro.

### Accessing built-in help

Command: `help`

This command will show the help information about `proot-distro` usage.
* `proot-distro help` - main page.
* `proot-distro <command> --help` - view help for specific command.

### Backing up distribution

Command: `backup`

Backup specified distribution and its plug-in into tar archive. The contents
of backup can be either printed to stdout for further processing or written
to a file.

Compression is determined according to file extension, e.g.`.tar.gz` will lead
to GZip compression and `.tar.xz` will lead to XZ. Piped backup data is always
not compressed giving user freedom for further processing.

Usage example:
```
proot-distro backup debian | xz | ssh example.com 'cat > /backups/pd-debian-backup.tar.xz'
proot-distro backup --output backup.tar.gz debian
```

*This command is generic. All additional processing like encryption should be
done by user through external commands.*

### Installing a distribution

Command: `install`

Install a distribution specified by alias - a short name referring to the
plug-in of chosen distribution.

Usage example:
```
proot-distro install alpine
```

By default the installed distribution will have same alias as specified on
command line. This means you will be unable to install multiple copies at
same time. You can rename distribution during installation time by using
option `--override-alias` which will create a copy of distribution plug-in.

Usage example:
```
proot-distro install --override-alias alpine-test alpine
proot-distro login alpine-test
```

Copied plug-in has following name format `<name>.override.sh` and is stored
in directory with others (`$PREFIX/etc/proot-distro`).

### Listing distributions

Command: `list`

Shows a list of available distributions, their aliases, installation status
and comments.

### Start shell session

Command: `login`

Execute a shell within the given distribution. Example:
```
proot-distro login debian
```

Execute a shell as specified user in the given distribution:
```
proot-distro login --user admin debian
```

You can run a custom command as well:
```
proot-distro login debian -- /usr/local/bin/mycommand --sample-option1
```

Argument `--` acts as terminator of `proot-distro login` options processing.
All arguments behind it would not be treated as options of PRoot Distro.

Login command supports these behavior modifying options:
* `--user <username>`

  Use a custom login user instead of default `root`. You need to create the
  user via `useradd -U -m username` before using this option.

* `--fix-low-ports`

  Force redirect low networking ports to a high number (2000 + port). Use
  this with software requiring low ports which are not possible without real
  root permissions.

  For example this option will redirect port 80 to something like 2080.

* `--isolated`

  Do not mount host volumes inside chroot environment. If this option was
  given, following mount points will not be accessible inside chroot:

  * /apex (only Android 10+)
  * /data/dalvik-cache
  * /data/data/com.termux
  * /sdcard
  * /storage
  * /system
  * /vendor

  You will not be able to use Termux utilities inside chroot environment.

* `--termux-home`

  Mount Termux home directory as user home inside chroot environment.

  This option takes priority over option `--isolated`.

* `--shared-tmp`

  Share Termux temporary directory with chroot environment. Takes priority
  over option `--isolated`.

* `--bind path:path`

  Create a custom file system path binding. Option expects argument in the
  given format:
  ```
  <host path>:<chroot path>
  ```

  Takes priority over option `--isolated`.

* `--no-link2symlink`

  Disable PRoot link2symlink extension. This will disable hard link emulation.
  You can use this option only if SELinux is disabled or is in permissive mode.

* `--no-sysvipc`

  Disable PRoot System V IPC emulation. Try this option if you experience
  crashes.

* `--no-kill-on-exit`

  Do not kill processes when shell session terminates. Typically will cause
  session to hang if you have any background processes running.

### Uninstall distribution

Command: `remove`

This command completely deletes the installation of given system. Be careful
as it does not ask for confirmation. Deleted data is irrecoverably lost.

Usage example:
```
proot-distro remove debian
```

### Reinstall distribution

Command: `reset`

Delete the specified distribution and install it again. This is a shortcut for
```
proot-distro remove <dist> && proot-distro install <dist>
```

Usage example:
```
proot-distro reset debian
```

Same as with command `remove`, deleted data is lost irrecoverably. Be careful.

### Restore from backup

Command: `restore`

Restore the distribution from the given proot-distro backup (tar archive).

Restore operation performs a complete rollback to the backup state as was in
archive. Be careful as this command deletes previous data irrecoverably.

Compression is determined automatically from file extension. Piped data
must be always uncompressed before being supplied to `proot-distro`.

Usage example:
```
ssh example.com 'cat /backups/pd-debian-backup.tar.xz' | xz -d | proot-distro restore
proot-distro restore ./pd-debian-backup.tar.xz
```

### Clear downloads cache

Command: `clear-cache`

This will remove all cached root file system archives. 

## Adding distribution

Distribution is defined through the plug-in script that contains variables
with metadata. A minimal one would look like this:
```.bash
DISTRO_NAME="Debian"
TARBALL_URL['aarch64']="https://github.com/termux/proot-distro/releases/download/v1.10.1/debian-aarch64-pd-v1.10.1.tar.xz"
TARBALL_SHA256['aarch64']="f34802fbb300b4d088a638c638683fd2bfc1c03f4b40fa4cb7d2113231401a21"
```

Script is stored in directory `$PREFIX/etc/proot-distro` and should be named
like `<alias>.sh`, where `<alias>` is a desired name for referencing the
distribution. For example, Debian plug-in will typically be named `debian.sh`.

### Plug-in variables reference

`DISTRO_ARCH`: specifies which CPU architecture variant of distribution to
install.

Normally this variable is determined automatically and you should not set it.
Typical use case is to set a custom architecture to run the distribution under
QEMU emulator (user mode).

Supported architectures are: `aarch64`, `arm`, `i686`, `x86_64`.

`DISTRO_NAME`: a name of distribution, something like "Alpine Linux (3.14.1)".

`DISTRO_COMMENT`: comments for current distribution.

Normally this variable is not needed. Use it to notify user that something is
not working or additional steps required to get started with this distribution.

`TARBALL_STRIP_OPT`: how many leading path components should be stripped when
extracting rootfs archive. The default value is 1 because all default rootfs
tarballs store contents in a sub directory.

`TARBALL_URL`: a Bash associative array of root file system tarballs URLs.

Should be defined at least for your CPU architecture. Valid architecture names
are same as for `DISTRO_ARCH`.

`TARBALL_SHA256`: a Bash associative array of SHA-256 checksums for each rootfs
variant.

Must be defined for each tarball set in `TARBALL_URL`.

### Running additional installation steps

Plug-in can be configured to execute specified commands after installing the
distribution. This is done through function `distro_setup`.

Example:
```.bash
distro_setup() {
	run_proot_cmd apt update
	run_proot_cmd apt upgrade -yq
}
```

`run_proot_cmd` is used when command should be executed inside the rootfs.

## Differences from Chroot

While PRoot is often referred as userspace chroot implementation, it is much
different from it both by implementation and features of work. Here is a list
of most significant differences you should be aware of.

1. PRoot is slow.

   Every process is hooked through `ptrace()`, so PRoot can hijack the system
   call arguments and return values. This is typically used to translate file
   paths so traced program will see the different file system layout.

2. PRoot cannot detach from the running process.

   Since PRoot controls the running processes via `ptrace()` it cannot detach
   from them. This means you can't start a daemon process (e.g. sshd) and close
   PRoot session. You will have to either kill process, wait until it finish or
   let proot kill it immediately on session close.

3. PRoot does not elevate privileges.

   Chroot also does not elevate privileges on its own. Just PRoot is configured
   to hijack user id as well, i.e. make it appear as `root`. So in reality your
   user name, id and privileges remain to be same as without PRoot but programs
   that doing sanity check for current user will assume you are running as
   root user.

   Particularly, the fake root user makes it possible to use package manager
   in chroot environment.

[Termux]: <https://termux.com>
[proot]: <https://github.com/termux/proot>
