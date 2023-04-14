# DNS Provider Design

The DNS provider supports resources that perform DNS updates and data sources for reading DNS information.

Below we have a collection of _Goals_ and _Patterns_: they represent the guiding principles applied during the
development of this provider. Some are in place, others are ongoing processes, others are still just inspirational.

## Goals

* [_Stability over features_](.github/CONTRIBUTING.md)
* Support resources that update `A`, `AAAA`, `CNAME`, `MX`, `NS`, `PTR`, `SRV`, and `TXT` record types.
* Support data sources that read `A`, `AAAA`, `CNAME`, `MX`, `NS`, `PTR`, `SRV`, and `TXT` record types.
* Support configuring secret key based transaction authentication ([RFC 2845](https://datatracker.ietf.org/doc/html/rfc2845)) or GSS-TSIG ([RFC 3645](https://datatracker.ietf.org/doc/html/rfc3645))
* Provide comprehensive documentation 
* Highlight intended and unadvisable usages

The DNS provider supports resources that perform DNS updates ([RFC 2136](https://datatracker.ietf.org/doc/html/rfc2136)) and data sources for reading DNS information. The provider can be configured with secret key based transaction authentication ([RFC 2845](https://datatracker.ietf.org/doc/html/rfc2845)) or GSS-TSIG ([RFC 3645](https://datatracker.ietf.org/doc/html/rfc3645)).


## Patterns

General to development:

* **Avoid repetition**: the entities managed can sometimes require similar pieces of logic and/or schema to be realised.
  When this happens it's important to keep the code shared in communal sections, so to avoid having to modify code in
  multiple places when they start changing.
* **Test expectations as well as bugs**: While it's typical to write tests to exercise a new functionality, it's key to
  also provide tests for issues that get identified and fixed, so to prove resolution as well as avoid regression.
* **Automate boring tasks**: Processes that are manual, repetitive and can be automated, should be. In addition to be a
  time-saving practice, this ensures consistency and reduces human error (ex. static code analysis).
* **Semantic versioning**: Adhering to HashiCorp's own
  [Versioning Specification](https://www.terraform.io/plugin/sdkv2/best-practices/versioning#versioning-specification)
  ensures we provide a consistent practitioner experience, and a clear process to deprecation and decommission.