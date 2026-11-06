package agent

import "github.com/Autumn-27/norma/llm"

// cloneToolMessages copies the message slice and each content slice so a
// projection can rewrite a block without mutating the SDK working set.
func cloneToolMessages(req llm.CompletionRequest) llm.CompletionRequest {
	req.Messages = append([]llm.Message(nil), req.Messages...)
	for i := range req.Messages {
		req.Messages[i].Content = append([]llm.ContentBlock(nil), req.Messages[i].Content...)
	}
	return req
}
