package milieu

import "errors"

var ErrPSQLNotActive = errors.New("psql is not enabled in this Milieu instance")
