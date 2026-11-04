package agent

import (
	"context"
	"encoding/json"

	actool "github.com/Autumn-27/norma/tool"
)

// 减代码：取消工具输出预算。
//
// 原实现把单次工具输出限制在 30000 字符（context_get 每项默认 8000、最多 24000），
// 超出直接返回错误“引用或字段路径元数据超出输出上限”，大结果读不全、要反复分段。
// 这里保留函数名与签名（十余处调用点不动），但不再设上限。

const toolOutputUnlimited = 1 << 30

func readOutputBudget(context.Context) int { return toolOutputUnlimited }

func readContentBudget(_ context.Context, requested, _ int) int { return requested }

func contextJSON(_ context.Context, v any) (actool.Result, error) {
	b, e := json.Marshal(v)
	if e != nil {
		return actool.Errorf(e.Error()), nil
	}
	return actool.Text(string(b)), nil
}
