# APT provisioner for Packer

This plugin simplifies installing packages with APT while provisioning images
with [Packer](https://www.packer.io/).

## Synopsis

```
  "provisioners": [
    {
      "type": "apt",
      "packages": ["less", "vim-tiny"]
    }
```

## Setup

Prerequisites:
- `golang-go` to build this plugin from source
- `packer` (use `--no-install-recommends` to prevent the Debian package of
  Packer from also installing Docker when you don't need it)
- (optional)
  [packer-builder-nspawn-debootstrap](https://git.sr.ht/~angdraug/packer-builder-nspawn-debootstrap/)
  to build Debian container images with systemd-nspawn

For compatibility with the Debian package of Packer that is built with a newer
version of [ugorji-go-codec](https://github.com/ugorji/go) than the one pinned
in Packer source, this plugin's `go.mod` includes a replace line to import a
similarly patched version of Packer source. To use this plugin with Packer
built from unpatched upstream source, comment out that replace line.

## Configuration

- `packages` - list of packages to install. The plugin uses
  `--no-install-recommends` and will not install recommended packages that are
  not explicitly enumerated.

- `sources` - additional APT sources to be listed under
  `/etc/apt/sources.list.d`.

- `keys` - list of files with public OpenPGP keys to be used for authenticating
  packages from the additional APT sources. The key files will be placed under
  `/etc/apt/trusted.gpg.d` and should use either .gpg (`gpg --export`) or .asc
  (`gpg --export --armor`) format as expected by
  [apt-secure(8)](https://manpages.debian.org/unstable/apt/apt-secure.8.en.html).

- `cache_dir` - local APT cache directory. The default is
  `/var/cache/apt/archives`. The contents will be copied into the image under
  `/var/cache/apt/archives` before running `apt-get install`, and will be
  purged from the image with `apt-get clean` afterwards.

When building multiple images with the same or overlapping set of packages, you
can pre-populate APT cache before running packer:

```
apt-get install -y -d -o Dir::Cache::Archives=${cache_dir} ${packages}
```

The `Dir::Cache::Archives` part of this command is only necessary if you want
to keep your package cache on the host separate from APT's default
`/var/cache/apt/archives` and pass a non-default value of `cache_dir` to the
provisioner.

## Copying

Copyright (c) 2020  Dmitry Borodaenko <angdraug@debian.org>

This program is free software. You can distribute/modify this program under
the terms of the GNU General Public License version 3 or later, or under
the terms of the Mozilla Public License, v. 2.0, at your discretion.

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

This Source Code Form is not "Incompatible With Secondary Licenses",
as defined by the Mozilla Public License, v. 2.0.
