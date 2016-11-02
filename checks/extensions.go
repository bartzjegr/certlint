package checks

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"strings"
	"sync"

	"github.com/globalsign/certlint/certdata"
)

var extMutex = &sync.Mutex{}

type extensions []extensionCheck

type extensionCheck struct {
	name   string
	oid    asn1.ObjectIdentifier
	filter *Filter
	f      func(pkix.Extension, *certdata.Data) []error
}

// Extensions contains all imported extension checks
var Extensions extensions

// RegisterExtensionCheck adds a new check to Extensions
func RegisterExtensionCheck(name string, oid asn1.ObjectIdentifier, filter *Filter, f func(pkix.Extension, *certdata.Data) []error) {
	extMutex.Lock()
	Extensions = append(Extensions, extensionCheck{name, oid, filter, f})
	extMutex.Unlock()
}

// Check lookups the registered extension checks and runs all checks with the
// same Object Identifier.
func (e extensions) Check(ext pkix.Extension, d *certdata.Data) []error {
	var errors []error
	var found bool

	for _, ec := range e {
		if ec.oid.Equal(ext.Id) {
			found = true
			if ec.filter != nil && ec.filter.Check(d) {
				continue
			}
			errors = append(errors, ec.f(ext, d)...)
		}
	}

	if !found {
		// Don't report private enterprise extensions as unknown, registered private
		// extensions have still been checked above.
		if !strings.HasPrefix(ext.Id.String(), "1.3.6.1.4.1.") {
			errors = append(errors, fmt.Errorf("Certificate contains unknown extension (%s)", ext.Id.String()))
		}
	}

	return errors
}