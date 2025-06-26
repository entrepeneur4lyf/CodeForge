#!/usr/bin/env python3
import sys
import json
import traceback
from sentence_transformers import SentenceTransformer

# Load the model2vec distilled model
MODEL_PATH = "/tmp/codeforge-model-1695886931/minilm-distilled"

try:
    model = SentenceTransformer(MODEL_PATH)
except Exception as e:
    print(json.dumps({"error": f"Failed to load model: {str(e)}"}))
    sys.exit(1)

def generate_embedding(text):
    try:
        # Generate embedding
        embedding = model.encode(text, convert_to_tensor=False)
        
        # Convert to list for JSON serialization
        embedding_list = embedding.tolist()
        
        return {"embedding": embedding_list}
    except Exception as e:
        return {"error": f"Failed to generate embedding: {str(e)}"}

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print(json.dumps({"error": "Usage: script.py <text>"}))
        sys.exit(1)
    
    text = sys.argv[1]
    result = generate_embedding(text)
    print(json.dumps(result))
