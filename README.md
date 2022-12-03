# Script to format a commit message

## Requirements

Requires:

- git
- xargs
- grep

## Compilation

Tested with golang v1.17, compile with:

```
make
```

## Usage

Once compiled, move `commit-message` in your $PATH and copy
`prepare-commit-msg` in the `.git/hooks/` directory of your project.

## Features

- Provides commit type
- Provides string or int types of fields
- Looks at the list of staged files to figure out which package to display in
the message, if there are more than one, the user will be prompted to choose at
least one.
- Looks at the branch name to try to pre-determine a Jira ticket number to use.
For it, the branch name must have the format `[A-Z]+-\d+*` (must start with the
ticket number)

## Showcase

```
> git commit
Select the commit type:
  1 - code
  2 - fix
  3 - chore
  4 - refactor
  5 - test
  6 - build
  7 - doc
  8 - nit
  9 - tool
 10 - remove
 11 - infra
 12 - hint
(1-12)> 2
Jira ticket ABC-585 found, use it?
(yes/no)> no
Commit message: Some message
[ABC-585-feat/great-change] fix(some/package) - Some message
 1 file changed, 1 insertion(+), 1 deletion(-)
```
