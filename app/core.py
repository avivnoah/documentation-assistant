import streamlit as st
import requests
import os
from typing import Set

GO_SERVER_URL = os.getenv("GO_LLM_URL", "http://localhost:8080")

st.header("Documentation Assistant")

if (
    "user_prompt_history" not in st.session_state
    and "chat_answers_history" not in st.session_state
    and "chat_history" not in st.session_state
):
    st.session_state["user_prompt_history"] = []
    st.session_state["chat_answers_history"] = []
    st.session_state["chat_history"] = []


# Add tabs for different functionalities
tab1, tab2 = st.tabs(["Query", "Ingest New Docs"])


def create_sources_string(source_urls: Set[str]) -> str:
    if not source_urls:
        return ""
    sources_list = list(source_urls)
    sources_list.sort()
    sources_string = "sources:\n"
    for i, source in enumerate(sources_list):
        sources_string += f"{i+1}. {source}\n"
    return sources_string


# Tab 1: Query Documentation
with tab1:
    st.subheader("Ask Questions")
    prompt = st.text_input("Prompt", placeholder="Enter your prompt here...", key="query_input")

    def call_go_llm(query: str, chat_history: list) -> dict:
        try:
            payload = {"query": query, "num_docs": 5, "chat_history": chat_history}
            resp = requests.post(f"{GO_SERVER_URL}/run", json=payload, timeout=60)
            resp.raise_for_status()
            return resp.json()
        except Exception as e:
            return {"error": str(e)}

    if prompt:
        with st.spinner("Generating response..."):
            generated_response = call_go_llm(
                query=prompt,
                chat_history=st.session_state["chat_history"],
            )
            if generated_response.get("error"):
                st.error("Error from Go server: " + generated_response["error"])
            else:
                print(generated_response.keys())
                sources = set(
                    [doc["Metadata"]["source"] for doc in generated_response["source_documents"]]
                )

                formatted_response = (
                    generated_response["result"] + "\n\n" + create_sources_string(sources)
                )

                st.session_state["user_prompt_history"].append(prompt)
                st.session_state["chat_answers_history"].append(formatted_response)
                st.session_state["chat_history"].append(("human", prompt)) # (Role, content)
                st.session_state["chat_history"].append(("ai", generated_response["result"])) # (Role, content)


                if st.session_state["chat_answers_history"]:
                    for generated_response, user_query in zip(
                        st.session_state["chat_answers_history"],
                        st.session_state["user_prompt_history"],
                    ):
                        st.chat_message("user").write(user_query)
                        st.chat_message("assistant").write(generated_response)

# Tab 2: Ingest New Documentation
with tab2:
    st.subheader("Ingest New Documentation")
    st.write("Enter a URL to crawl and add to the knowledge base.")
    
    new_url = st.text_input("Documentation URL", placeholder="https://example.com/docs", key="ingest_url")
    
    def ingest_documentation(url: str) -> dict:
        try:
            payload = {"url": url}
            resp = requests.post(f"{GO_SERVER_URL}/ingest", json=payload, timeout=10)
            resp.raise_for_status()
            return resp.json()
        except Exception as e:
            return {"error": str(e)}
    
    if st.button("Start Ingestion", type="primary"):
        if new_url:
            with st.spinner("Starting ingestion process..."):
                result = ingest_documentation(new_url)
                if result.get("error"):
                    st.error("Error: " + result["error"])
                else:
                    st.success(result.get("message", "Ingestion started!"))
                    st.info("The ingestion process is running in the background. This may take several minutes depending on the documentation size.")
        else:
            st.warning("Please enter a URL to ingest.")
