package trigger

import (
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/runtime"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

type XpertServiceBinding interface {
	Dic() *di.Container
	Runtime() *runtime.GolangRuntime
}
