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

Then create a `.git/commit_template` file in your repository.
This file must contain the commit format. For example:

```
{t} {i;Year} - {m}
```

Will lead to the following when committing:

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
  8 - tool
  9 - remove
 10 - infra
 11 - hint
(1-11)> 2
Year: 1970
Commit message: Some message
[ABC-585-feat/great-change] fix 1970 - Some message
 1 file changed, 1 insertion(+), 1 deletion(-)
```

The following templates are possible:
- `{t}`: Commit type (code, fix, chore, refactor...)
- `{p}`: Relevant packages (affected directories)
- `{n}`: Ticket number (in the `ABC-123` format)
- `{m}`: Commit message
- `{i;some prompt}`: Integer
- `{s;some prompt}`: Arbitrary string

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
  8 - tool
  9 - remove
 10 - infra
 11 - hint
(1-11)> 2
Jira ticket ABC-585 found, use it?
(yes/no)> no
Commit message: Some message
[ABC-585-feat/great-change] fix(some/package) - Some message
 1 file changed, 1 insertion(+), 1 deletion(-)
```
