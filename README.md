# grpc-ipfs-lite

A gRPC wrapper around [ipfs-lite](https://github.com/hsanjuan/ipfs-lite)

### IPFS-lite Libraries

> The following includes information about support for ipfs-lite.

| Name | Build | Language | Description |
|:---------|:---------|:---------|:---------|
| [`ipfs-lite`](https://github.com/hsanjuan/ipfs-lite) | [![Build Status](https://img.shields.io/travis/hsanjuan/ipfs-lite.svg?branch=master&style=flat-square)](https://travis-ci.org/hsanjuan/ipfs-lite) | [![golang](https://img.shields.io/badge/golang-blueviolet.svg?style=popout-square)](https://github.com/hsanjuan/ipfs-lite) | The reference implementaiton of ipfs-lite, written in Go. |
| [`js-ipfs-lite`](//github.com/textileio/js-ipfs-lite) | [![Build status](https://img.shields.io/github/workflow/status/textileio/js-ipfs-lite/Test/master.svg?style=popout-square)](https://github.com/textileio/js-ipfs-lite/actions?query=branch%3Amaster) | [![javascript](https://img.shields.io/badge/javascript-blueviolet.svg?style=popout-square)](https://github.com/textileio/js-ipfs-lite)| The Javascript version of ipfs-lite available for web, nodejs, and React Native applications. |
| [`ios-ipfs-lite`](//github.com/textileio/ios-ipfs-lite) | [![Build status](https://img.shields.io/circleci/project/github/textileio/ios-ipfs-lite/master.svg?style=flat-square)](https://github.com/textileio/ios-ipfs-lite/actions?query=branch%3Amaster) | [![objc](https://img.shields.io/badge/objc-blueviolet.svg?style=popout-square)](https://github.com/textileio/ios-ipfs-lite)| The iOS ipfs-lite library for use in Objc and Swift apps |
| [`android-ipfs-lite`](//github.com/textileio/android-ipfs-lite) | [![Build status](https://img.shields.io/circleci/project/github/textileio/android-ipfs-lite/master.svg?style=flat-square)](https://github.com/textileio/android-ipfs-lite/actions?query=branch%3Amaster) | [![java](https://img.shields.io/badge/java-blueviolet.svg?style=popout-square)](https://github.com/textileio/android-ipfs-lite)| The Java ipfs-lite library for us in Android apps |
| [`grpc-ipfs-lite`](//github.com/textileio/grpc-ipfs-lite) | [![Build status](https://img.shields.io/circleci/project/github/textileio/grpc-ipfs-lite/master.svg?style=flat-square)](https://github.com/textileio/grpc-ipfs-lite/actions?query=branch%3Amaster) | [![java](https://img.shields.io/badge/grpc--api-blueviolet.svg?style=popout-square)](https://github.com/textileio/grpc-ipfs-lite)| A common gRPC API interface that runs on the Go ipfs-lite node. |

## What is IPFS Lite?

From the [ipfs-lite](https://github.com/hsanjuan/ipfs-lite) project:

IPFS-Lite is an embeddable, lightweight IPFS peer which runs the minimal setup to provide an ipld.DAGService. It can:

    Add, Get, Remove IPLD Nodes to/from the IPFS Network (remove is a local blockstore operation).
    Add single files (chunk, build the DAG and Add) from a io.Reader.
    Get single files given a their CID.

It provides:

    An ipld.DAGService
    An AddFile method to add content from a reader
    A GetFile method to get a file from IPFS.

## What is the gRPC wrapper?

[gRPC](https://grpc.io/) is a modern open source high performance RPC framework that can run in any environment.

This project adds a minimal gRPC service on top of the IPFS Lite module. It allows you to embed IPFS Lite into multiple projects while exposing a common API. You could deploy IPFS Lite as a microservice, embed it in mobile applications, or wrap it in a Dockerfile.

## Projects using gRPC IPFS Lite

* [Android IPFS Lite](https://github.com/textileio/android-ipfs-lite)
* [iOS IPFS Lite](https://github.com/textileio/ios-ipfs-lite)

**PR your own project link to the list above**
