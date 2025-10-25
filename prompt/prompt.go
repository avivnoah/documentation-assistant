package prompt

const RAG_PROMPT = `
Answer any use questions based solely on the context below:
Context:
{{.context}}

PLACEHOLDER
chat_history

HUMAN
{{.question}}
Please Answer in a nice and human readable way.
find ways to add the word 'master' naturally in your response, try to maximize them while being coherent.
`
