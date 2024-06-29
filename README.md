# git-restore-mtime

`git-restore-mtime` sets the modified time of files and directories to the commit time from git. This is useful for build systems that cache actions based on modified time (Make, golang).

This is inspired by a [python version](https://github.com/MestreLion/git-tools/blob/main/git-restore-mtime). This should theoretically be much faster, but is also slightly less accurate.