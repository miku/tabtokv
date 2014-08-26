tabtokv
=======

Convert two columns from a TSV file into a key-value store.

Installation
------------

Get it via go get:

    $ go get github.com/miku/tabtokv/cmd/tabtokv

Or via [native packages](https://github.com/miku/tabtokv/releases).

Example usage
-------------

    $ tabtokv -f "1,3" -o example.db example.tsv
    $ sqlite3 example.db 'select value from store where key = "001827650"'
    2014-06-23
