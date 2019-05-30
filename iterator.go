package rovers

import (
	"bufio"
	"io"
	"os"
)

// OrganizationIterator iters returning a github organization time each time.
type OrganizationIterator interface {
	// Next returns the next github organization. If the iterator is already
	// cosumed it will return io.EOF
	Next() (string, error)
	// Close closes the iterator
	Close() error
	// ForEach applies the given function to all the github organizations the
	// iterator contains.
	ForEach(func(org string) error) error
}

func forEachOrgIterator(iter OrganizationIterator, fn func(string) error) error {
	defer iter.Close()
	for {
		org, err := iter.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}

			return err
		}

		if err := fn(org); err != nil {
			return err
		}
	}
}

// orgsIterFile use a file with a github organization name per line to iter.
type orgsIterFile struct {
	r       io.ReadCloser
	scanner *bufio.Scanner
}

// NewOrganizationIterator builds a new OrganizationIterator.
func NewOrganizationIterator(path string) (OrganizationIterator, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return &orgsIterFile{
		r:       f,
		scanner: bufio.NewScanner(f),
	}, nil
}

// Next implements the OrganizationIterators interface.
func (i *orgsIterFile) Next() (string, error) {
	if !i.scanner.Scan() {
		err := i.scanner.Err()
		if err == nil {
			return "", io.EOF
		}
	}

	org := i.scanner.Text()
	if org == "" {
		return i.Next()
	}

	return org, nil
}

// Close implements the OrganizationIterators interface.
func (i *orgsIterFile) Close() error {
	return i.r.Close()
}

// ForEach implements the OrganizationIterators interface.
func (i *orgsIterFile) ForEach(fn func(string) error) error {
	return forEachOrgIterator(i, fn)
}

type orgsIterSlice struct {
	orgs []string
}

// NewOrgIterFromSlice builds a new OrganizationIterator.
func NewOrgIterFromSlice(orgs []string) OrganizationIterator {
	return &orgsIterSlice{orgs: orgs}
}

// Next implements the OrganizationIterators interface.
func (i *orgsIterSlice) Next() (string, error) {
	if len(i.orgs) == 0 {
		return "", io.EOF
	}

	var org string
	org, i.orgs = i.orgs[0], i.orgs[1:]
	return org, nil
}

// Close implements the OrganizationIterators interface.
func (i *orgsIterSlice) Close() error {
	return nil
}

// ForEach implements the OrganizationIterators interface.
func (i *orgsIterSlice) ForEach(fn func(string) error) error {
	return forEachOrgIterator(i, fn)
}
