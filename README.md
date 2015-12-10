[![Build Status](https://travis-ci.org/balzaczyy/golucene.svg?branch=master)](https://travis-ci.org/balzaczyy/golucene)
[![Coverage Status](https://coveralls.io/repos/balzaczyy/golucene/badge.png?branch=lucene410)](https://coveralls.io/r/balzaczyy/golucene?branch=lucene410)
[![GoDoc](https://godoc.org/github.com/balzaczyy/golucene?status.svg)](https://godoc.org/github.com/balzaczyy/golucene)

golucene
========

A [Go](http://golang.org) port of [Apache Lucene](http://lucene.apache.org). Check out the online demo [here](http://hamlet.zhouyiyan.cn/)!

Sync to Lucene Java 4.10 2014/10/22

Why do we need yet another port of Lucene?
------------------------------------------

Since Lucene Java is already optimized to teeth (and yes, I know it very much), potential performance gain should not be expected from its Go port. Quote from Lucy's FAQ:

>Is Lucy faster than Lucene? It's written in C, after all.

>That depends. As of this writing, Lucy launches faster than Lucene thanks to tighter integration with the system IO cache, but Lucene is faster in terms of raw indexing and search throughput once it gets going. These differences reflect the distinct priorities of the most active developers within the Lucy and Lucene communities more than anything else.

It also applies to GoLucene. But some benefits can still be expected:
- quick start speed;
- able to be embedded in Go app;
- goroutine which I think can be faster in certain case;
- ready-to-use byte, array utilities which can reduce the code size, and lead to easy maintenance.

Though it started as a pet project, I've been pretty serious about this.

Dependencies
------------
Go 1.2+

Installation
------------

	go get -u github.com/balzaczyy/golucene

Usage
-----

Using GoLucene is similar to using Lucene Java. Firstly, index need
to be built first. Then, create an query and do the search against
the index.

Note that the current GoLucene is rather basic and limited in feature.
Only default functions are supported, like term frequency based
weight calculation, filesystem directory, boolean query, etc. For
further features, please raise requests.

A detailed example can be found [here](gl.go).

License
-------
Apache Public License 2.0.
