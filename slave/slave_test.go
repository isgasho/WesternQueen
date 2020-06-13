package slave

import (
	"github.com/arcosx/WesternQueen/util"
	"testing"
)

func init() {
	util.DebugMode = true
	util.Mode = "slave1"
}
func TestStart(t *testing.T) {
	Start()
}
