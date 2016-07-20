package icinterface

import (
	"github.com/RexGene/common/timingwheel"
)

type ISender interface {
	AddTimer(cb func()) *timingwheel.BaseNode
	Resend(token uint, backupData interface{})
}
