import streamlit as st
import requests
import os

GO_SERVER_URL = os.getenv("GO_LLM_URL", "http://localhost:8080")

st.header("Documentation Assistant")

# Add tabs for different functionalities
tab1, tab2 = st.tabs(["Query", "Ingest New Docs"])

# Tab 1: Query Documentation
with tab1:
    st.subheader("Ask Questions")
    prompt = st.text_input("Prompt", placeholder="Enter your prompt here...", key="query_input")

    def call_go_llm(prompt_text: str) -> dict:
        try:
            payload = {"query": prompt_text, "num_docs": 5}
            resp = requests.post(f"{GO_SERVER_URL}/run", json=payload, timeout=60)
            resp.raise_for_status()
            return resp.json()
        except Exception as e:
            return {"error": str(e)}

    if prompt:
        with st.spinner("Generating response..."):
            result = call_go_llm(prompt)
            if result.get("error"):
                st.error("Error from Go server: " + result["error"])
            else:
                st.write("**Answer:**")
                st.write(result.get("result") or result)
                
                # Show source documents if available
                if result.get("source_documents"):
                    with st.expander("View Source Documents"):
                        st.json(result["source_documents"])

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