
INSTALL
=======

irdmtools is an **experimental** Go package and command line tools for working with institutional repositories (e.g. Invenio RDM). It is distributed in source code and in binary form for macOS (Intel and M1), Linux (Intel and ARM 64), Raspberry Pi OS.

To test the latest version you need the required development environment and follow the steps listed below in "Compiling from Source".

Quick install using curl or irm
-------------------------------

The following experimental installer should get the latest stable release for macOS and Linux (e.g. Debian, Ubuntu, Raspberry Pi OS). 

Copy and run the following command in your shell (e.g. Terminal)

~~~
curl https://caltechlibrary.github.io/irdmtools/installer.sh | sh
~~~

For Windows you can use a Powershell script with the following command.

~~~
irm https://caltechlibrary.github.io/irdmtools/installer.ps1 | iex
~~~

If you want to install a specific version you can download the installer scripts. then pass the version on the command line. As an example to install version v0.0.83 specifically you'd type the following two command into your shell session.

~~~shell
curl https://caltechlibrary.github.io/irdmtools/installer.sh
sh installer.sh 0.0.83
~~~

Requirements
------------

This may change in the future.

- Git to clone the repository from GitHub
- [Golang](https://golang.org) 1.20.4 or better
- GNU Make
- Pandoc 3 or better (to build documentation)
- Bash 
- grep

Compiling from Source
---------------------

1. clone the repository
2. change into the cloned directory
3. run "make", "make test" and "make install"

Here's the steps I take to build and test on my macOS box or Linux box.

~~~
git clone git@github.com:caltechlibrary/irdmtools.git
cd irdmtools
make
make test
make install
~~~

