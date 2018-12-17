[![GoDoc](https://godoc.org/github.com/shuLhan/share?status.svg)](https://godoc.org/github.com/shuLhan/share)
[![Go Report Card](https://goreportcard.com/badge/github.com/shuLhan/share)](https://goreportcard.com/report/github.com/shuLhan/share)

A collection of libraries and tools written in Go.

## Command Line Interface

- `gofmtcomment`: Program to convert "/\*\*/" comment into "//".


## Libraries

- `bytes`: Package bytes provide a library for working with byte or slice of
  bytes.

- `contact`: Package contact provide a library to import contact from Google,
  Microsoft, and Yahoo.

- `dns`: Package dns implement DNS client and server.

- `dsv`: Package dsv is a library for working with delimited separated value
(DSV).

- `errors`: Package errors provide an error type with code.

- `git`: Package git provide a wrapper for git command line interface.

- `http`: Package http implement custom HTTP server with memory file system
and simplified routing handler.

- `ini`: Package ini implement reading and writing INI configuration as
defined by Git configuration file syntax.

- `io`: Package io provide a library for reading and watching file, and
reading from standard input.

- `memfs`: Package memfs provide a library for mapping file system into
memory.

- `mining`: Package mining provide a library for data mining.

  - `classifier/cart`: Package cart implement the Classification and
  Regression Tree by Breiman, et al.

  - `classifier/crf`: Package crf implement the cascaded random forest
  algorithm, proposed by Baumann et.al in their paper: Baumann, Florian, et
  al.

  - `classifier/rf`: Package rf implement ensemble of classifiers using
  random forest algorithm by Breiman and Cutler.

  - `gain/gini`: Package gini contain function to calculating Gini gain.

  - `knn`: Package knn implement the K Nearest Neighbor using Euclidian to
  compute the distance between samples.

  - `math`: Package math provide generic functions working with mathematic.

  - `resampling/lnsmote`: Package lnsmote implement the Local-Neighborhood
  algorithm from the paper, Maciejewski, Tomasz, and Jerzy Stefanowski.

  - `resampling/smote`: Package smote resamples a dataset by applying
  the Synthetic Minority Oversampling TEchnique (SMOTE).

  - `tree/binary`: Package binary contain implementation of binary tree.

- `net`: Package net provide constants and library for networking.

- `numbers`: Package numbers provide miscellaneous functions for working with
integer, float, slice of integer, and slice of floats.

- `runes`: Package runes provide a library for working with a single rune or
slice of rune.

- `strings`: Package string provide a library for working with string or slice
of string.

- `tabula`: Package tabula is a Go library for working with rows, columns, or
matrix (table), or in another terms working with data set.

- `test`: Package test provide library for help with testing.

- `text`: Package text provide common a library for working with text.
  - `diff`: Package diff implement text comparison.

- `time`: Package time provide a library for working with time.

- `websocket`: Package websocket provide the websocket library for server
and client.
