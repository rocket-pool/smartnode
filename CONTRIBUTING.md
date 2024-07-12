# Contributing to Smart Node
Smart Node is an open-source project that welcomes community contributions.

## Roles
Smart Node has three roles: Owner, Maintainers, and Contributors.

### Owner
The Owner has final say in settling disputes and making executive decisions when consensus cannot be reached. Currently, the Owner is @0xfornax.

### Maintainers
Maintainers are team or community members who contribute routinely. The Owner is also a Maintainer. Current Maintainers:
- @0xfornax
- @thomaspanf
- @jshufro

Smart Node was previously maintained by @jclapis and @moles1

### Contributors
Contributors are community members who have submitted merged pull requests with some regularity. They are too numerous to list individually.

## Good Practices
- Adhere to golang best practices. See [Effective Go](https://golang.org/doc/effective_go.html) for a good starting point.
- Adhere to broader coding practices, such as DRY, POLA, Avoid Deep Nesting, etc.
- New code should be unit tested when possible.
- Write self-documenting code when possible, and add comments when necessary.

## Pull Requests
- Each PR should represent a single feature, bugfix, or non-functional change.
- Owners and Maintainers may decide if a change is too small for peer review, but should still offer the opportunity for review by posting the PR for at least a day, unless urgent.
- Larger changes require at least one Maintainer's review.
- All Contributor PRs must be reviewed by a Maintainer.
- Commits should be groomed: each commit should compile, pass tests, and represent logical atomic progressions.
- History rewriting is permissible only while a PR is in DRAFT status, or after a Maintainer has approved the PR. Once a PR is approved, fixup commits should be rebased into the original commits.

## Licenses
Smart Node is licensed under GNU GPLv3, and all contributions must be made under the same license. See [LICENSE](LICENSE) for details.