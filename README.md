[![Build Status](https://travis-ci.org/balzaczyy/golucene.svg?branch=master)](https://travis-ci.org/balzaczyy/golucene)

golucene
========

A [Go](http://golang.org) port of [Apache Lucene](http://lucene.apache.org). Check out the online demo [here](http://hamlet.mybluemix.net/)!

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

	import (
	  "fmt"
		std "github.com/balzaczyy/golucene/analysis/standard"
		_ "github.com/balzaczyy/golucene/core/codec/lucene410"
		docu "github.com/balzaczyy/golucene/core/document"
		"github.com/balzaczyy/golucene/core/index"
		"github.com/balzaczyy/golucene/core/search"
		"github.com/balzaczyy/golucene/core/store"
		"github.com/balzaczyy/golucene/core/util"
	)

	util.SetDefaultInfoStream(util.NewPrintStreamInfoStream(os.Stdout))
	index.DefaultSimilarity = func() index.Similarity {
		return search.NewDefaultSimilarity()
	}
	
	...

	directory, _ := store.OpenFSDirectory("app/index")
	analyzer := std.NewStandardAnalyzer()
	conf := index.NewIndexWriterConfig(util.VERSION_LATEST, analyzer)
	writer, _ := index.NewIndexWriter(directory, conf)

	d := docu.NewDocument()
	d.Add(docu.NewTextFieldFromString("foo", "bar", docu.STORE_YES))
	writer.AddDocument(d.Fields())
	writer.Close() // ensure index is written

	reader, _ := index.OpenDirectoryReader(directory)
	searcher := search.NewIndexSearcher(reader)
	
	q := search.NewTermQuery(index.NewTerm("foo", "bar"))
	res, _ := searcher.Search(q, nil, 1000)
	fmt.Printf("Found %v hits.\n", res.TotalHits)
	for _, hit := range res.ScoreDocs {
	  fmt.Printf("Doc %v score: %v\n", hit.Doc, hit.Score)
	}

License
-------
Apache Public License 2.0.
