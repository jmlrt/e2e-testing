# Formal verification of Metricbeat module

We want to make sure that a metricbeat module works as expected, taking into
consideration different versions of the integration software that metricbeats uses.
So for that reason we are adding [smoke tests](http://softwaretestingfundamentals.com/smoke-testing/) to verify that each module, and different versions for that module, meets to the specifications described here with certain grade of satisfaction.

>Smoke Testing, also known as “Build Verification Testing”, is a type of software testing that comprises of a non-exhaustive set of tests that aim at ensuring that the most important functions work. The result of this testing is used to decide if a build is stable enough to proceed with further testing.

## Running the tests

The tests are located under the [root](./) directory. Place your terminal there and execute the Godog tests, which is the official Golang implementation for Cucumber:

```shell
$ GO111MODULE=on go test -v --godog.format pretty [apache|filebeat|helm|metricbeat|mysql|redis|vsphere]
```

## Tooling

The specification of these smoke tests has been done using the `BDD` (Behaviour-Driven Development) principles, where:

>BDD aims to narrow the communication gaps between team members, foster better understanding of the customer and promote continuous communication with real world examples.

The implementation of these smoke tests has been done with [Godog](https://github.com/cucumber/godog) + [Cucumber](https://cucumber.io/).

### Cucumber: BDD at its core

From their website:

>Cucumber is a tool that supports Behaviour-Driven Development(BDD), and it reads executable specifications written in plain text and validates that the software does what those specifications say. The specifications consists of multiple examples, or scenarios.

The way we are going to specify our software is using [`Gherkin`](https://cucumber.io/docs/gherkin/reference/).

>Gherkin uses a set of special keywords to give structure and meaning to executable specifications. Each keyword is translated to many spoken languages. Most lines in a Gherkin document start with one of the keywords.

The key part here is **executable specifications**: we will be able to automate the verification of the spefications anf potentially get a coverage of these specs.

### Godog: Cucumber for Golang

From their website:

>Package godog is the official Cucumber BDD framework for Golang, it merges specification and test documentation into one cohesive whole.

For this POC, we have chosen Godog over any other test framework because the team is using already using Golang, so it seems reasonable to choose it.

## Test Specification

All the Gherkin (Cucumber) specifications are written in `.feature` files.

A good example could be [this one](./features/metricbeat/mysql.feature):

```cucumber
Feature: As a Metricbeat developer I want to check that the MySQL module works as expected

Scenario Outline: Check module is sending metrics to Elasticsearch without errors
  Given "<variant>" v<version>, variant of "MySQL", is running for metricbeat
    And metricbeat is installed and configured for "<variant>", variant of the "MySQL" module
    And metricbeat waits "20" seconds for the service
  When metricbeat runs for "20" seconds
  Then there are "<variant>" events in the index
    And there are no errors in the index
Examples:
| variant | version    |
| MariaDB | 10.2.23    |
| MariaDB | 10.3.14    |
| MariaDB | 10.4.4     |
| MySQL   | 5.7.12     |
| MySQL   | 8.0.13     |
| Percona | 5.7.24     |
| Percona | 8.0.13-4   |
```

## Test Implementation

We are using Godog + Cucumber to implement the tests, where we create connections to the `Given`, `When`, `Then`, `And`, etc. in a well-known file structure.

As an example, the Golang implementation of the `features/metricbeat/mysql.feature` is located under the [./metricbeat_test.go](./metricbeat_test.go) file.

Each module will define its own file for specificacions, adding specific feature context functions that will allow filtering the execution, if needed. These functions would be managed by a map in the test runner. We see this as a workaround to be improved in a future refactor, but we had to continue the development of the PoC while Godog solves the coupling of the life cycle hooks for each test suite, as demonstrated [here](https://github.com/mdelapenya/sample-godog). 

```shell
# Will run all scenarios for mysql
$ GO111MODULE=on go test --godog.format pretty mysql
```