bettercsv [![Build Status](https://travis-ci.org/tinygrasshopper/bettercsv.svg)](https://travis-ci.org/tinygrasshopper/bettercsv)
=========

A drop in replacement for the stdlib csv implementation for golang, to make it support any [dsv](http://en.wikipedia.org/wiki/Delimiter-separated_values). Fully backward compatible while providing more features.

Forked from the 'encoding/csv' golang(1.3.3) csv implementation. 

# Usage
In addition to [encoding/csv](http://golang.org/pkg/encoding/csv/) functionality, you can specify custom quoting characters for reading and writing. 

## Example
```
fileReader, _ = os.Open("/tmp/wierd_dsv.txt")
csvReader := bettercsv.NewReader(fileReader)
csvReader.Comma = ';'
csvReader.Quote = '|'
content := csvReader.ReadAll()
```


# Install

`go get github.com/tinygrasshopper/bettercsv`


# Goals
- [x] Custom quote rune while reading
- [x] Custom quote rune while writing
- [x] Support no quoting charater while writing
- [ ] Supporting headers
- [ ] Reading to a map
