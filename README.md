tabtokv
=======

Convert two columns from a TSV file into a key-value store.

Example usage
-------------

    $ tabtokv -f "1,3" -o example.db example.tsv
    $ sqlite3 example.db 'select value from store where key = "001827650"'
    2014-06-23
