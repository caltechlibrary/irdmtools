INSTALL
=======

irdmtools is an experimental Go package. It is distributed in source code and in binary form for macOS (Intel and M1), Linux (Intel and ARM 64), Raspberry Pi OS. 

Quick install using curl
------------------------

The following experimental installer should work for macOS and Linux
(e.g. Debian, Ubuntu, Raspberry Pi OS)

Copy and run the following command in your shell (e.g. Terminal)

~~~
curl https://caltechlibrary.github.io/datatools/installer.sh | sh
~~~

Installing from source
----------------------

If you wish to compile irdmtools and test it you need the required development environment and follow the steps listed below in "Compiling from Source".

Requirements
------------

This may change in the future.

- Git to clone the repository
- [Golang](https://golang.org) 1.19.2 or better
- GNU Make
- Pandoc 2.19.2 or better (to build documentation)
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

