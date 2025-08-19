# Regex Tables

## About the project

This project is a library for creating and working with regex tables in Go.
A regex table is a kind of associative array whose keys are regular
expressions (regexs) that map to arbitrary values. A lookup operations matches
a strings against the regex-keys to find a match and returns the corresponding
value _and_ the match groups.

Core to the implementation is the compilation of the regex-keys into a single
regular expression with named capture groups for each key. This allows the
lookup operation to efficiently match the input string against the combined 
regex and extract the corresponding value for the matched key.

This package abstracts over different regex engines. We can use a 
RegexTableBuilder to construct regex tables efficiently.

## Temporary Files

- VSCode gets confused by temporary files too easily. And when you try 
  deleted them it often instantly recreates them. So you must always check
  that they are gone after a few seconds have elapsed (i.e. add sleep to
  the command).

- When you need to create temporary files, avoid creating them in the repo 
  folder - unless it is in the `tmp` folder, which is excluded by .gitignore.
  It is fine to create them in `/tmp` too.

## Programming Guidelines

- Comments should be proper sentences, with correct grammar and punctuation,
  including the use of capitalization and periods.

- Where defensive checks are added, include a comment explaining why they are
  appropriate (not necessary, since defensive checks are not necessary).


