# Terraform Provider: DNS

The DNS provider supports resources that perform DNS updates ([RFC 2136](https://datatracker.ietf.org/doc/html/rfc2136)) and data sources for reading DNS information. The provider can be configured with secret key based transaction authentication ([RFC 2845](https://datatracker.ietf.org/doc/html/rfc2845)) or GSS-TSIG ([RFC 3645](https://datatracker.ietf.org/doc/html/rfc3645)).

## Requirements

* [Terraform](https://www.terraform.io/downloads)
* [Go](https://go.dev/doc/install) (1.22)
* [GNU Make](https://www.gnu.org/software/make/)
* [golangci-lint](https://golangci-lint.run/usage/install/#local-installation) (optional)

## Documentation, questions and discussions
Official documentation on how to use this provider can be found on the
[Terraform Registry](https://registry.terraform.io/providers/hashicorp/DNS/latest/docs).
In case of specific questions or discussions, please use the
HashiCorp [Terraform Providers Discuss forums](https://discuss.hashicorp.com/c/terraform-providers/31),
in accordance with HashiCorp [Community Guidelines](https://www.hashicorp.com/community-guidelines).

We also provide:

* [Support](.github/SUPPORT.md) page for help when using the provider
* [Contributing](.github/CONTRIBUTING.md) guidelines in case you want to help this project

## Compatibility

Compatibility table between this provider, the [Terraform Plugin Protocol](https://www.terraform.io/plugin/how-terraform-works#terraform-plugin-protocol)
version it implements, and Terraform:

| DNS Provider | Terraform Plugin Protocol | Terraform |
|:------------:|:-------------------------:|:---------:|
|  `>= 3.0.x`  |            `5`            | `>= 0.12` |
|  `>= 2.1.x`  |        `4` and `5`        | `>= 0.12` |
|  `<= 2.x.x`  |            `4`            | `<= 0.12` |

Details can be found querying the [Registry API](https://www.terraform.io/internals/provider-registry-protocol#list-available-versions)
that return all the details about which version are currently available for a particular provider.
[Here](https://registry.terraform.io/v1/providers/hashicorp/DNS/versions) are the details for DNS (JSON response).


## Development

### Building

1. `git clone` this repository and `cd` into its directory
2. `make` will trigger the Golang build

The provided `GNUmakefile` defines additional commands generally useful during development,
like for running tests, generating documentation, code formatting and linting.
Taking a look at it's content is recommended.

### Testing

In order to test the provider, you can run

* `make test` to run provider tests
* `make testacc` to run provider acceptance tests (excluding ones requiring a `DNS_UPDATE_SERVER`)
* `./internal/provider/acceptance.sh` to run the full suite of acceptance tests

Running `acceptance.sh` has the following prerequisites:

- [Docker](https://www.docker.com/)
- [Go](https://golang.org/)
- [Kerberos Clients](https://web.mit.edu/kerberos/dist/) (e.g. `kinit`)
- [Make](https://www.gnu.org/software/make/)
- [Terraform CLI](https://terraform.io/)
- `/etc/hosts` entry (or equivalent): `127.0.0.1 ns.example.com`



### macOS Setup

- [Docker for Mac](https://docs.docker.com/docker-for-mac/install/)
- [Go](https://golang.org/dl/) or with Homebrew: `brew install go`
- [Terraform CLI](https://www.terraform.io/downloads.html) or with Homebrew: `brew install hashicorp/tap/terraform`

```shell
echo "127.0.0.1 ns.example.com" | sudo tee -a /etc/hosts
```

### Ubuntu Setup

- [Docker Engine](https://docs.docker.com/engine/install/ubuntu/)
- [Go](https://github.com/golang/go/wiki/Ubuntu)
- [Terraform CLI](https://www.terraform.io/docs/cli/install/apt.html)

```shell
echo "127.0.0.1 ns.example.com" | sudo tee -a /etc/hosts
sudo apt-get install krb5-user make
# If prompted for Kerberos configuration:
# Default Realm: EXAMPLE.COM
# Server: ns.example.com
# Administrative Server: ns.example.com
```

It's important to note that acceptance tests (`testacc`) will actually spawn
`terraform` and the provider. Read more about they work on the
[official page](https://www.terraform.io/plugin/testing/acceptance-tests).

### Generating documentation

This provider uses [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs/)
to generate documentation and store it in the `docs/` directory.
Once a release is cut, the Terraform Registry will download the documentation from `docs/`
and associate it with the release version. Read more about how this works on the
[official page](https://www.terraform.io/registry/providers/docs).

Use `make generate` to ensure the documentation is regenerated with any changes.

### Using a development build

If [running tests and acceptance tests](#testing) isn't enough, it's possible to set up a local terraform configuration
to use a development builds of the provider. This can be achieved by leveraging the Terraform CLI
[configuration file development overrides](https://www.terraform.io/cli/config/config-file#development-overrides-for-provider-developers).

First, use `make install` to place a fresh development build of the provider in your
[`${GOBIN}`](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies)
(defaults to `${GOPATH}/bin` or `${HOME}/go/bin` if `${GOPATH}` is not set). Repeat
this every time you make changes to the provider locally.

Then, setup your environment following [these instructions](https://www.terraform.io/plugin/debugging#terraform-cli-development-overrides)
to make your local terraform use your local build.

### Testing GitHub Actions

This project uses [GitHub Actions](https://docs.github.com/en/actions/automating-builds-and-tests) to realize its CI.

Sometimes it might be helpful to locally reproduce the behaviour of those actions,
and for this we use [act](https://github.com/nektos/act). Once installed, you can _simulate_ the actions executed
when opening a PR with:

```shell
# List of workflows for the 'pull_request' action
$ act -l pull_request

# Execute the workflows associated with the `pull_request' action 
$ act pull_request
```

## Releasing

The release process is automated via GitHub Actions, and it's defined in the Workflow
[release.yml](./.github/workflows/release.yml).

Each release is cut by pushing a [semantically versioned](https://semver.org/) tag to the default branch.

## License

[Mozilla Public License v2.0](./LICENSE)
