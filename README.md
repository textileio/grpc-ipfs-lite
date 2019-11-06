# grpc-ipfs-lite

A gRPC wrapper around [ipfs-lite](https://github.com/hsanjuan/ipfs-lite)

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