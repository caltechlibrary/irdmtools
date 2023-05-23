#!/bin/sh

#
# Set the package name and version to install
#
PACKAGE="irdmtools"
# NOTE: GitHub prefixes versions with 'v' in paths, we'll do that too.
VERSION="v0.0.12"
RELEASE="https://github.com/caltechlibrary/$PACKAGE/releases/tag/$VERSION"

#
# Get the name of this script.
#
INSTALLER="$(basename "$0")"

#
# Figure out what the zip file is named
#
OS_NAME="$(uname -o)"
MACHINE="$(uname -m)"
case "$OS_NAME" in
   Darwin)
   OS_NAME="macOS"
   ;;
   "GNU/Linux")
   OS_NAME="Linux"
   ;;
esac
ZIPFILE="$PACKAGE-$VERSION-$OS_NAME-$MACHINE.zip"
echo "Downloading $ZIPFILE"

#
# Check to see if this zip file has been downloaded.
#
DOWNLOAD_URL="https://github.com/caltechlibrary/$PACKAGE/releases/download/$VERSION/$ZIPFILE"
if ! curl -L -o "$HOME/Downloads/$ZIPFILE" "$DOWNLOAD_URL"; then
	echo "Curl failed to get $DOWNLOAD_URL"
fi
cat<<EOT

  Retrieved $DOWNLOAD_URL
  Saved as $HOME/Downloads/$ZIPFILE

EOT

if [ ! -f "$HOME/Downloads/$ZIPFILE" ]; then
	cat<<EOT

  To install $PACKAGE you need to download 

    $ZIPFILE

  from 

    $RELEASE

  You can do that with your web browser. After
  that you should be able to re-run $INSTALLER

EOT
	exit 1
fi

START="$(pwd)"
mkdir -p "$HOME/.$PACKAGE/installer"
cd "$HOME/.$PACKAGE/installer" || exit 1
unzip "$HOME/Downloads/$ZIPFILE" "bin/*" "man/*"

#
# Copy the application into place
#
mkdir -p "$HOME/bin"
EXPLAIN_OS_POLICY="yes"
# NOTE: writing to prog_names.tmp to meet avoid pipeline.
find bin -type f >prog_names.tmp
while read -r APP; do
	V=$("./$APP" --version)
	if [ "$V" = ""  ]; then 
		EXPLAIN_OS_POLICY="yes"
	fi
	mv "$APP" "$HOME/bin/"
done <prog_names.tmp
rm prog_names.tmp

#
# Make sure $HOME/bin is in the path
#
DIR_IN_PATH='no'
for P in $PATH; do
  if [ "$P" = "$HOME/bin" ]; then
	 DIR_IN_PATH='yes'
  fi
done
if [ "$DIR_IN_PATH" = "no" ]; then
	# shellcheck disable=SC2016
	echo 'export PATH="$HOME/bin:$PATH"' >>"$HOME/.bashrc"
	# shellcheck disable=SC2016
	echo 'export PATH="$HOME/bin:$PATH"' >>"$HOME/.zshrc"
fi
rm -fR "$HOME/.$PACKAGE/installer"
cd "$START" || exit 1

# shellcheck disable=SC2031
if [ "$EXPLAIN_OS_POLICY" = "no" ]; then
	cat <<EOT

  You need to take additional steps to complete installation.

  Your operating system security policied needs to "allow" 
  running programs from $PACKAGE.

  Example: on macOS you can type open the programs in finder.

      open $HOME/bin

  Find the program(s) and right click on the program(s)
  installed to enable them to run.

EOT

fi

#
# Copy the manual pages into place
#
EXPLAIN_MAN_PATH="no"
for SECTION in 1 2 3 4 5 6 7; do
	if [ -f "man/man${SECTION}" ]; then
		EXPLAIN_MAN_PATH="yes"
		mkdir -p "$HOME/man/man${SECTION}"
		find "man/man${SECTION}" -type f | while read -r MAN; do
			cp -v "$MAN" "$HOME/man/man${SECTION}/"
		done
	fi
done

if [ "$EXPLAIN_MAN_PATH" = "yes" ]; then
  cat <<EOT
  The man pages have been installed at '$HOME/man'. You
  need to have that location in your MANPATH for man to
  find the pages. E.g. For the Bash shell add the
  following to your following to your '$HOME/.bashrc' file.

      export MANPATH="$HOME/man:$MANPATH"
  
EOT

fi
