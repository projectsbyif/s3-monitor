# Monitoring data within Amazon S3 with Trillian #

This repository is a combination of Go code used in Amazon Lambda functions and
Terraform code to deploy the functions including an example application using
CloudTrail.

This project deploys [Trillian](https://github.com/google/trillian) using
Amazon's Lambda and Aurora serverless services to produce verifiable datasets
based on data written into Amazon S3.

## Why might this be useful ##

Data used for many services running in/on Amazon Web Services (AWS) is stored
using Amazon's object store S3. In many cases this data is sensitive, and the
accuracy and integrity of the data is important.

Trillian is a tool which produces transparent verifiable datasets, used in
applications such as Certificate Transparency. Using Trillian it's possible to
provide proofs that data has been added in an "append-only" fashion. The log is
said to be tamper-evident as any edits to elements already included in the
Merkle tree can be identified.

These proofs can be used in ways without revealing the underlying data. For
instance in cases where sensitive data is held about a user they could acquire a
proof data about them is held without seeing data about other users.

## How it works ##

Objects added to an S3 bucket trigger a Lambda function which adds metadata
about the new object into Trillian. This produces a new leaf in the Merkle Tree
recorded by Trillian.

Using this Merkle Tree data structure Trillian is able produce efficient proofs
about the integrity of the data stored within.

Currently our code includes new leaves into the Merkle tree once a day. The new
signed log root is written to a separate S3 bucket which could be monitored for
consistency.

![AWS architecture diagram](./images/aws-architecture.svg)

### Deploying ###

Requirements: a working [Go programming environment](https://golang.org/doc/install), [Terraform](https://www.terraform.io/).

```shell
  # go get a bunch of dependencies - https://coderwall.com/p/arxtja/install-all-go-project-dependencies-in-one-command
  go get ./...
  # build the lambda functions
  make
  # deploy !
  cd terraform
  terraform apply
```

## Example use case: CloudTrail ##

Amazon Web Services includes CloudTrail, an audit log of all activity within the
account. These log files are delivered into Amazon's object store S3.

The terraform code included here will set up CloudTrail logging into the bucket
monitored by Trillian. This serves as a convenient example of how monitoring
datasets in S3 could work.

### CloudTrail's existing integrity checks ###

While CloudTrail includes existing integrity checks, they don't allow for the
same types of proofs provided by Trillian. However they work with existing
tools and work well, and if you use AWS you should enable them. This project is
not an alternative.

### What can be proven with Trillian ###

Each day a new Signed Log Root is published making it possible to continuously
monitor the "append-only" nature of the log by asking Trillian to issue a
consistency proof.

This consistency proof includes hashes of intermediary leaves added in the past
day. Using these you are able to establish that the previous log root has been
appended to, not edited.

## Limitations ##

This is exploratory work and shouldn't seen as a replacement for the integrity
support which exists in AWS services like CloudTrail already.

While Trillian itself uses secure hashing functions when creating leaves
(typically SHA-256) by following rfc6962. Currently S3 events only include a
hash of the file in the way of an eTag (intended for cache busting) which uses
md5, a famously broken cryptographic hash function.

It would be great if AWS began supplying a modern cryptographically secure hash
of the object in the events captured by s3-monitor.

## Other use cases ##

In some cases, it will make sense to use other features in Trillian. For
instance, if you're sharing data into an S3 bucket and require proof your data
has been received as it was sent Trillian may provide an "inclusion" proof.
There may also be similar use cases in proving data is absent or has been
removed from a dataset.

We have an upcoming blog post which will cover possible future applications of
this technology.
