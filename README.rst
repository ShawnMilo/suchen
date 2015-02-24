======
suchen
======

``suchen`` is meant to be a drop-in replacement for
ack-grep (http://beyondgrep.com/) for situations when you are unable
to install software on your server. Because it's written in Go, suchen
can be compiled to a stand-alone, static binary.

Case-insensitive matching can be done by passing ``-i`` as an argument.

Like ack-grep, an advantage over basic grep is that you can search only
within files with specific extensions. All searches are recursive.

``suchen`` should run faster than ack-grep on multi-core machines because it checks
files concurrently.

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
