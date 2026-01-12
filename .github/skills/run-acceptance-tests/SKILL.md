---
name: run-acceptance-tests
description: Guide for running acceptance tests for the DNS provider. Use this when asked to run an acceptance test or to run a test with the prefix `TestAcc`.
---

An acceptance test is a Go test function with the prefix `TestAcc`.

To run a focussed acceptance test named `TestAccFeatureHappyPath`:

1. Run `go test -run=TestAccFeatureHappyPath` with the following environment
   variables:
   - `TF_ACC=1`
   
   Default to non-verbose test output.

To diagnose a failing acceptance test, use these options, in order. These
options are cumulative: each option includes all the options above it.

1. Run the test again. Use the `-count=1` option to ensure that `go test` does
   not use a cached result.
1. Offer verbose `go test` output. Use the `-v` option.
1. Offer debug-level logging. Enable debug-level logging with the environment
   variable `TF_LOG=debug`.
1. Offer to persist the acceptance test's Terraform workspace. Enable
   persistance with the environment variable `TF_ACC_WORKING_DIR_PERSIST=1`.

To flip a passing acceptance test named `TestAccFeatureHappyPath`:

1. Edit the value of one of the TestCheckFuncs in one of the TestSteps in the
   TestCase.
1. Run the acceptance test. Expect the test to fail.
1. If the test fails, then undo the edit and report a successful flip. Else,
   keep the edit and report an unsuccessful flip.
