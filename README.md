# go-whosonfirst-iterate-github

Go package for iterating through a set of Who's On First documents stored in a GitHub repository, using the GitHub API.

## Important

Documentation for this package is incomplete and will be updated shortly.

## Example

```
package main

import (
       "context"
       "flag"
       "io"
       "log"

       _ "github.com/whosonfirst/go-whosonfirst-iterate-github/v2"
       
       "github.com/whosonfirst/go-whosonfirst-iterate/emitter/v2"       
       "github.com/whosonfirst/go-whosonfirst-iterate/indexer/v2"
)

func main() {

	emitter_uri := flag.String("emitter-uri", "githubapi://", "A valid whosonfirst/go-whosonfirst-iterate/emitter URI")
	
     	flag.Parse()

	ctx := context.Background()

	emitter_cb := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) error {
		log.Printf("Indexing %s\n", path)
		return nil
	}

	iter, _ := iterator.NewIterator(ctx, *emitter_uri, cb)

	uris := flag.Args()
	iter.IterateURIs(ctx, uris...)
}
```

_Error handling removed for the sake of brevity._


## URIs and Schemes 

### githubapi://

```
githubapi://{GITHUB_ORGANIZATION}/{GITHUB_REPO}
```

## Query parameters

In addition to the [default go-whosonfirst-iterate query parameters](https://github.com/whosonfirst/go-whosonfirst-iterate#query-parameters) the following query parameters are supported:

| Name | Value | Required | Notes
| --- | --- | --- | --- |
| access_token | String | Yes | A valid [GitHub API access token](https://docs.github.com/en/rest/overview/other-authentication-methods) |
| branch | String | No | The branch to use when iterating the repository contents |
| concurrent | Bool | No | If true iterate through documents concurrently. There is still a throttle on the number of API requests per second but this can speed things up significantly with the risk that you will still trigger GitHub API limits. |

## Filters

### QueryFilters

You can also specify inline queries by appending one or more `include` or `exclude` parameters to a `emitter.Emitter` URI, where the value is a string in the format of:

```
{PATH}={REGULAR EXPRESSION}
```

Paths follow the dot notation syntax used by the [tidwall/gjson](https://github.com/tidwall/gjson) package and regular expressions are any valid [Go language regular expression](https://golang.org/pkg/regexp/). Successful path lookups will be treated as a list of candidates and each candidate's string value will be tested against the regular expression's [MatchString](https://golang.org/pkg/regexp/#Regexp.MatchString) method.

For example:

```
repo://?include=properties.wof:placetype=region
```

You can pass multiple query parameters. For example:

```
repo://?include=properties.wof:placetype=region&include=properties.wof:name=(?i)new.*
```

The default query mode is to ensure that all queries match but you can also specify that only one or more queries need to match by appending a `include_mode` or `exclude_mode` parameter where the value is either "ANY" or "ALL".

## Tools

```
$> make cli
go build -mod vendor -o bin/count cmd/count/main.go
go build -mod vendor -o bin/emit cmd/emit/main.go
```

### count

Count files in one or more whosonfirst/go-whosonfirst-iterate/emitter sources.

```
> ./bin/count -h
Count files in one or more whosonfirst/go-whosonfirst-iterate/emitter sources.
Usage:
	 ./bin/count [options] uri(N) uri(N)
Valid options are:

  -emitter-uri string
        A valid whosonfirst/go-whosonfirst-iterate/emitter URI. Supported emitter URI schemes are: directory://,featurecollection://,file://,filelist://,geojsonl://,githubapi://,repo://

```

For example:

```
$> ./bin/count \
	-emitter-uri 'githubapi://sfomuseum-data/sfomuseum-data-architecture?concurrent=1&access_token={TOKEN}' \
	data

2021/03/02 13:06:08 time to index paths (1) 1m11.522679037s
2021/03/02 13:06:08 Counted 1077 records (1077) in 1m11.522714392s
```

Or:

```
$> ./bin/count \
	-emitter-uri 'githubapi://sfomuseum-data/sfomuseum-data-architecture?concurrent=1&access_token={TOKEN}&include=properties.sfomuseum:placetype=museum' \
	data

2021/03/02 13:35:25 time to index paths (1) 1m10.897179298s
2021/03/02 13:35:25 Counted 7 records (7) in 1m10.897222091s
```

### emit

Publish features from one or more whosonfirst/go-whosonfirst-index/v2/emitter sources.

```
> ./bin/emit -h
Publish features from one or more whosonfirst/go-whosonfirst-iterate/emitter sources.
Usage:
	 ./bin/emit [options] uri(N) uri(N)
Valid options are:

  -emitter-uri string
        A valid whosonfirst/go-whosonfirst-iterate/emitter URI. Supported emitter URI schemes are: directory://,featurecollection://,file://,filelist://,geojsonl://,githubapi://,repo://
  -geojson
    	Emit features as a well-formed GeoJSON FeatureCollection record.
  -json
    	Emit features as a well-formed JSON array.
  -null
    	Publish features to /dev/null
  -stdout
    	Publish features to STDOUT. (default true)
```

For example:

```
$> ./bin/emit \
	-emitter-uri 'githubapi://sfomuseum-data/sfomuseum-data-architecture?concurrent=1&access_token={TOKEN}&include=properties.sfomuseum:placetype=museum' \
	-geojson \	
	data

| jq '.features[]["properties"]["wof:id"]'

1729813675
1477855937
1360521563
1360521569
1360521565
1360521571
1159157863
```

## See also

* https://github.com/whosonfirst/go-whosonfirst-iterate
* https://github.com/google/go-github/github
* https://docs.github.com/en/rest/reference/repos#contents