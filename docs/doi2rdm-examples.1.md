%doi2rdm-examples(1) irdmtools user manual | version 0.1.0 6ccc3239
% R. S. Doiel and Tom Morrell
% 2026-10-27

# NAME

doi2rdm-examples - Save the default YAML options to a file.

# SYNOPSIS

doi2rdm -help examples

# DESCRIPTION

Save the default YAML options to a file. You can customize this to match your
vocabulary requirements in your RDM deployment.

~~~
	doi2rdm -show-yaml >options.yaml
~~~

Example generating a JSON document for a single DOI. The resulting
text file is called "article.json". In this example "options.yaml"
is the configuration file for setup for your RDM instance. It'll first
check CrossRef then DataCite.

~~~
	doi2rdm options.yaml "10.1021/acsami.7b15651" >article.json
~~~

See what CrossRef returned for a DOI, and what doi2rdm made of it. The raw
CrossRef JSON and any warnings are in "debug.txt", the RDM record is in
"article.json".

~~~
	doi2rdm -debug options.yaml "10.1021/acsami.7b15651" >article.json 2>debug.txt
~~~

Check to see the difference from the saved "article.json" and
the current metadata retrieved from CrossRef or DataCite.

~~~
	doi2rdm -diff article.json options.yaml "10.1021/acsami.7b15651"
~~~

Example getting metadata for an arXiv record from DataCite

~~~
	doi2rdm options.yaml "arXiv:2312.07215"
~~~

# SEE ALSO

doi2rdm(1)
