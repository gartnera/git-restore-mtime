# git-restore-mtime

`git-restore-mtime` sets the modified time of files and directories to the commit time from git.

This is inspired by a [python version](https://github.com/MestreLion/git-tools/blob/main/git-restore-mtime). This was supposed to be faster, but the python version is actually faster because it just parses the output of `git log`.

## Known Issues

- [worktree's are not supported](https://github.com/go-git/go-git/issues/483)
