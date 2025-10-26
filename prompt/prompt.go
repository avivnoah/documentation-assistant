package prompt

const RAG_PROMPT = `
Answer any use questions based solely on the context below:
Context:
{{.context}}

PLACEHOLDER
chat_history

HUMAN
{{.question}}
`

const REPHRASE_PROMPT = `
Given the following conversation and a follow up question, rephrase the follow up question to be a standalone question.

Chat History:
{{.chat_history}}
Follow Up Input: {{.question}}
Standalone Question:
`
