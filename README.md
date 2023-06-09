# CloudQuery cloudformation-template-summary Source Plugin

[![test](https://github.com/guardian/cq-source-cloudformation-template-summary/actions/workflows/test.yaml/badge.svg)](https://github.com/guardian/cq-source-cloudformation-template-summary/actions/workflows/test.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/guardian/cq-source-cloudformation-template-summary)](https://goreportcard.com/report/github.com/guardian/cq-source-cloudformation-template-summary)

A cloudformation-template-summary source plugin for CloudQuery that loads data from cloudformation-template-summary to any database, data warehouse or data lake supported by [CloudQuery](https://www.cloudquery.io/), such as PostgreSQL, BigQuery, Athena, and many more.

## Running this project
```
go run main.go serve
```

Then, once you have cloudquery installed (`brew install cloudquery/tap/cloudquery`), and the correct Janus credentials,

```
cloudquery sync templates.yaml
```

## Links

 - [CloudQuery Quickstart Guide](https://www.cloudquery.io/docs/quickstart)
 - [Supported Tables](docs/tables/README.md)


## Configuration

The following source configuration file will sync to a PostgreSQL database. See [the CloudQuery Quickstart](https://www.cloudquery.io/docs/quickstart) for more information on how to configure the source and destination.

```yaml
kind: source
spec:
  name: "cloudformation-template-summary"
  path: "guardian/cloudformation-template-summary"
  version: "${VERSION}"
  destinations:
    - "postgresql"
  spec:
    # plugin spec section
```

## Development

### Run tests

```bash
make test
```

### Run linter

```bash
make lint
```

### Generate docs

```bash
make gen-docs
```

### Release a new version

1. Run `git tag v1.0.0` to create a new tag for the release (replace `v1.0.0` with the new version number)
2. Run `git push origin v1.0.0` to push the tag to GitHub

Once the tag is pushed, a new GitHub Actions workflow will be triggered to build the release binaries and create the new release on GitHub.
To customize the release notes, see the Go releaser [changelog configuration docs](https://goreleaser.com/customization/changelog/#changelog).
