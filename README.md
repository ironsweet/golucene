golucene
========

A [Go](http://golang.org) port of [Apache Lucene](http://lucene.apache.org).

Why do we need yet another port of Lucene?

Since Lucene Java is already optimized to teeth (and yes, I know it very much), potential performance gain should not be expected from its Go port. Quote from Lucy's FAQ:

>Is Lucy faster than Lucene? It's written in C, after all.

>That depends. As of this writing, Lucy launches faster than Lucene thanks to tighter integration with the system IO cache, but Lucene is faster in terms of raw indexing and search throughput once it gets going. These differences reflect the distinct priorities of the most active developers within the Lucy and Lucene communities more than anything else.

It also applies to GoLucene. But some benefits can still be expected:
- quick start speed;
- able to be embedded in Go app;
- goroutine which I think can be faster in certain case;
- ready-to-use byte, array utilities which can reduce the code size, and lead to easy maintenance.

Though it started as a pet project, I've been pretty serious about this.

Milestones
----------
- 2013/6/3    Initial code commited.
- 2013/11/11  First test case (gl.go) that search for specific keyword has passed with number of hits.
- 2013/12/3   Second test case (gl.go) that fetch string field has passed.
- 2014/6/29   Third test case (gl3.go) that index string field has passed.
- 2014/8/8    Migrated to Lucene 4.9.0 code base.
- 2014/9/5    QueryParser scenario for simple keywords can be used.

Progress
--------
V1.0 ETA: 2014/12/1

TODOs
-----
- QueryParser.
- Finish fourth test case (gl2.go) with Lucene's test framework.
- Support basic explain function.

License
-------
Until further notice, all the unreleased codes and documents are licensed under APL 2.0.

Develop Guidelines
------------------
- NewXXX method can return value or pointer. Value is for simple data object, while pointer is for more complex action object or interface when you want to protect its internal fields.
- XXXReader and XXXDirectory must be pointers as they make changes to their underlying data structure.
- Parameter in pointer form indicates mutability.
- Use interface/inheritance when children don't add new capabilities.
- Use embeded interface/struct when children add new capabilities and explicitly used.
- Prefer value internally unless (a) optional field; or (b) obvious readonly fields.
- "not implemented yet" means in scope but not yet translated from Java.
- "not supported yet" means not in scope but will be implemented later.
- "not supported" equals to UnSupportedOperationException.
- Wrapper can be implemented by composition instead of an explicit delegate field.
- Methods start with '_' are considered "internal" and vulnerable to thread-safety. They must be invoked by either other "internal" methods or synchronized methods. This is to workaround universally used re-entrant lock offered by Java synchronized keyword.

Known Risks/Issues
------------------
- Volatile keyword is not supported yet.
