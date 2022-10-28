INSTALL
=======

irdmtools is an experimental Go package. It is only distributed in source code. If you wish to compile irdmtools and test it you need the required development environment and follow the steps listed below in "Compiling from Source".

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

