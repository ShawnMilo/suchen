======
suchen
======

``suchen`` is a search script for finding text very quickly within files. It
is intended to be used similarly to ack-grep, but should run faster. 
Case-insensitive matching can be done by passing ``-i`` as the first 
command-line argument.

Like ack-grep, an advantage over basic grep is that you can search only
within files with specific extensions. All searches are recursive.

Basic usage::

    # Find pattern "foo" in all files, starting in the current dir.
    suchen foo

Search in certain file extensions::

    # Find pattern "foo" in Python files, starting in the current dir.
    suchen foo --py

Using regular expressions::

    # Find imports of jQuery within HTML pages, regardless of version.
    suchen 'script.*jquery\.js' --html

Installation
============

None. Just copy the binary for your platform from the release, or compile
the source yourself with Go. It is recommended that you rename the binary
to just ``suchen``, since the downloads are named by platform.

License
=======

BSD
