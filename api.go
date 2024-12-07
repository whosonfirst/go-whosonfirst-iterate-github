package github

import (
	"context"
	"errors"
	_ "log"
	"net/url"
	"path/filepath"
	"strings"
	"strconv"
	"time"

	"github.com/google/go-github/v67/github"
	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3/iterator"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3/filters"
	"golang.org/x/oauth2"	
)

func init() {
	ctx := context.Background()
	iterator.RegisterIterator(ctx, "githubapi", NewGitHubAPIIterator)
}

type GitHubAPIIterator struct {
	iterator.Iterator
	owner string
	repo  string
	branch   string
	concurrent bool
	client   *github.Client
	throttle <-chan time.Time
	filters  filters.Filters	
}

func NewGitHubAPIIterator(ctx context.Context, uri string) (iterator.Iterator, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	rate := time.Second / 10
	throttle := time.Tick(rate)
	
	em := &GitHubAPIIterator{
		throttle: throttle,
	}

	em.owner = u.Host

	path := strings.TrimLeft(u.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) != 1 {
		return nil, errors.New("Invalid path")
	}

	em.repo = parts[0]
	em.branch = DEFAULT_BRANCH

	q := u.Query()

	token := q.Get("access_token")
	branch := q.Get("branch")
	concurrent := q.Get("concurrent")
	
	if token == "" {
		return nil, errors.New("Missing access token")
	}

	if branch != "" {
		em.branch = branch
	}

	if concurrent != "" {

		c, err := strconv.ParseBool(concurrent)

		if err != nil {
			return nil, err
		}

		em.concurrent = c
	}
	
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	em.client = client

	f, err := filters.NewQueryFiltersFromQuery(ctx, q)

	if err != nil {
		return nil, err
	}

	em.filters = f

	return em, nil
}

func (em *GitHubAPIIterator) Iterate(ctx context.Context, uris ...string) iter.Seq2[iterator.Record, error] {

	return func(yield func(Record, error) bool) {
		for _, uri := range uris {
			for r, err := em.iterate(ctx, uri) {
				yield(r, err)
			}
		}
	}
}

func (em *GitHubAPIIterator) iterate(ctx context.Context, uri string) iter.Seq2[iterator.Record, error] {

	return func(yield func(Record, error) bool) {

		// log.Printf("Walk %s/%s/%s", em.owner, em.repo, uri)
		
		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}
		
		file_contents, dir_contents, _, err := em.client.Repositories.GetContents(ctx, em.owner, em.repo, uri, nil)
		
		if err != nil {
			return yield(nil, err)
			return
		}

		if file_contents != nil {
			return em.walkFileContents(ctx, index_cb, file_contents)
		}
		
		if dir_contents != nil {

			var walk_iter iter.Seq2[iterator.Record, error]
			
			if em.concurrent {
				for r, err := range em.walkDirectoryContentsConcurrently(ctx, dir_contents) {
					yield(r, err)
				}
			} else {
				walk_rec, err := em.walkDirectoryContents(ctx, dir_contents)

				if err != nil {
					yield(nil, err)
				}

				if walk_rec != nil {
					yield(walk_rec, nil)
				}
			}

		}
		
	}
}

func (em *GitHubAPIIterator) walkDirectoryContents(ctx context.Context, index_cb iterator.IteratorCallbackFunc, contents []*github.RepositoryContent) error {

	for _, e := range contents {

		err := em.WalkURI(ctx, index_cb, *e.Path)
		
		if err != nil {
			return err
		}
	}

	return nil
}

func (em *GitHubAPIIterator) walkDirectoryContentsConcurrently(ctx context.Context, index_cb iterator.IteratorCallbackFunc, contents []*github.RepositoryContent) error {

	remaining := len(contents)

	done_ch := make(chan bool)
	err_ch := make(chan error)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, e := range contents {

		go func(e *github.RepositoryContent) {

			defer func() {
				done_ch <- true
			}()

			err := em.WalkURI(ctx, index_cb, *e.Path)

			if err != nil {
				err_ch <- err
			}
			
		}(e)
	}

	for remaining > 0 {
		select {
		case <-done_ch:
			remaining -= 1
		case err := <-err_ch:
			return err
		default:
			// pass
		}
	}

	return nil
}

func (em *GitHubAPIIterator) walkFileContents(ctx context.Context, index_cb iterator.IteratorCallbackFunc, contents *github.RepositoryContent) (iterator.Record, error) {

	path := *contents.Path
	name := *contents.Name

	switch filepath.Ext(name) {
	case ".geojson":
		// continue
	default:
		return nil, nil
	}

	body, err := contents.GetContent()

	if err != nil {
		return nil, err
	}

	r := strings.NewReader(body)

	fh, err := ioutil.NewReadSeekCloser(r)

	if err != nil {
		return nil, err
	}

	if em.filters != nil {

		ok, err := em.filters.Apply(ctx, fh)

		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, nil
		}

		_, err = fh.Seek(0, 0)

		if err != nil {
			return nil, err
		}
	}

	iter_rec := iterator.NewRecord(path, fh)
	return iter_rec, nil)
}
