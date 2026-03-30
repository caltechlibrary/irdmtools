Installation for development of **irdmtools**
===========================================

**irdmtools** Tools for working with institutional repositories and data management systems. Current implementation targets Invenio-RDM.

Quick install with curl or irm
------------------------------

There is an experimental installer.sh script that can be run with the following command to install latest table release. This may work for macOS, Linux and if you’re using Windows with the Unix subsystem. This would be run from your shell (e.g. Terminal on macOS).

~~~shell
curl https://caltechlibrary.github.io/irdmtools/installer.sh | sh
~~~

This will install the programs included in irdmtools in your `$HOME/bin` directory.

If you are running Windows 10 or 11 use the Powershell command below.

~~~ps1
irm https://caltechlibrary.github.io/irdmtools/installer.ps1 | iex
~~~

### If your are running macOS or Windows

You may get security warnings if you are using macOS or Windows. See the notes for the specific operating system you're using to fix issues.

- [INSTALL_NOTES_macOS.md](INSTALL_NOTES_macOS.md)
- [INSTALL_NOTES_Windows.md](INSTALL_NOTES_Windows.md)

Installing from source
----------------------

### Required software

- Go &gt;&#x3D; 1.26.1
- CMTools &gt;&#x3D; 0.0.40

### Suggested software

- PostgreSQL &gt;&#x3D; 16
- PostgREST &gt;&#x3D; 12
- Pandoc &gt;&#x3D; 3.1
- MySQL &gt;&#x3D; 8
- SQLite &gt;&#x3D; 3.49

### Steps

1. git clone https://github.com/caltechlibrary/irdmtools
2. Change directory into the `irdmtools` directory
3. Make to build, test and install

~~~shell
git clone https://github.com/caltechlibrary/irdmtools
cd irdmtools
make
make test
make install
~~~
