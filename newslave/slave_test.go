package newslave

import (
	"github.com/arcosx/WesternQueen/util"
	"testing"
)

func init() {
	util.DebugMode = true
	util.Mode = "slave1"

}

func Test_start(t *testing.T) {
	Start()
}