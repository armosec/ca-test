package elastic

import "fmt"

type ElasticServerOption func(opts *options, isUpdate bool) error

//Options for test server
type options struct {
	indexPrefix2Mapping map[string]string
	indexSuffix         string
}

//Options

//WithIndicesMapping option uses the provided mapping for indices instead of the default hard-coded mapping
var WithIndicesMapping = func(indexPrefix2Mapping map[string]string) ElasticServerOption {
	return func(o *options, isUpdate bool) error {
		if isUpdate {
			return fmt.Errorf("indices mapping can't be updated")
		}
		o.indexPrefix2Mapping = indexPrefix2Mapping
		return nil
	}
}

func makeOptions(opts ...ElasticServerOption) (*options, error) {
	o := &options{
		indexPrefix2Mapping: map[string]string{},
		indexSuffix:         "-9-2022",
	}
	return applyOptions(o, false, opts...)
}

func applyOptions(o *options, isUpdate bool, opts ...ElasticServerOption) (*options, error) {
	for _, option := range opts {
		if err := option(o, isUpdate); err != nil {
			return nil, err
		}
	}
	return o, nil
}
