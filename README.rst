======
suchen
======

``suchen`` is a search script for finding text very quickly within files.

Basic usage::

    # Find pattern "foo" in all files, starting in the current dir.
    suchen foo

Search in certain file extensions::

    # Find pattern "foo" in Python files, starting in the current dir.
    suchen foo --py

Using regular expressions::

    # Find imports of jQuery within HTML pages, regardless of version.
    suchen 'script.*jquery\.js' --html
