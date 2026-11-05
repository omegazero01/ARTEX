package server

import (
	"context"

	"github.com/neouks/series/db"
)

// plainFindingRecorder 是 agent 上报漏洞的入口。ARTEX 原版没有禁报规则，
// 这里拿到请求就直接写库：不匹配规则、不阻断、不写“禁报校验”活动记录。
type plainFindingRecorder struct{ s *Server }

func (s *Server) findingRecorder() *plainFindingRecorder { return &plainFindingRecorder{s: s} }

func (r *plainFindingRecorder) Record(ctx context.Context, in db.RecordFindingInput, refs []db.TrafficRef) (*db.RecordedFinding, error) {
	return r.s.evidenceStore().Record(ctx, in, refs)
}
