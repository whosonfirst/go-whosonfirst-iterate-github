package main

import (
	"context"
	"flag"
	"fmt"
	_ "github.com/whosonfirst/go-whosonfirst-iterate-github/v2"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/emitter"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"io"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

func main() {

	valid_schemes := strings.Join(emitter.Schemes(), ",")
	emitter_desc := fmt.Sprintf("A valid whosonfirst/go-whosonfirst-iterate/v2 URI. Supported emitter URI schemes are: %s", valid_schemes)

	var emitter_uri = flag.String("emitter-uri", "", emitter_desc)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Count files in one or more whosonfirst/go-whosonfirst-iterate/v2 sources.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t %s [options] uri(N) uri(N)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	ctx := context.Background()

	var count int64
	count = 0

	emitter_cb := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) error {

		atomic.AddInt64(&count, 1)
		return nil
	}

	iter, err := iterator.NewIterator(ctx, *emitter_uri, emitter_cb)

	if err != nil {
		log.Fatal(err)
	}

	paths := flag.Args()

	t1 := time.Now()

	err = iter.IterateURIs(ctx, paths...)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Counted %d records (%d) in %v\n", count, iter.Seen, time.Since(t1))
}
