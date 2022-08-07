package merror

import "github.com/hashicorp/go-multierror"

// MaybeMultiError only creates a multi error if there are multiple errors.
func MaybeMultiError(baseErr error, furtherErrs ...error) error {
	errOut := baseErr
	for _, e := range furtherErrs {
		if e != nil {
			errOut = multierror.Append(errOut, e)
		}
	}
	return errOut
}
