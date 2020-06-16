package newmaster

import (
	"github.com/arcosx/WesternQueen/util"
	"testing"
)

func TestSendShareWrongTraceSet(t *testing.T) {
	WrongTraceSet.Add("fuck you")
	WrongTraceSet.Add("memeda")
	SendShareWrongTraceSet(util.SLAVE_ONE_PORT_8000)
}