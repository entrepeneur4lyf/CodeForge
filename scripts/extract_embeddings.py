#!/usr/bin/env python3
"""
Extract embedding weights from model2vec for direct use in Go
"""

import os
import sys
import json
import numpy as np
from pathlib import Path
from sentence_transformers import SentenceTransformer
import torch

def extract_model_weights(model_path, output_path):
    """Extract embedding weights and tokenizer from model2vec"""
    
    print(f"Loading model from: {model_path}")
    
    # Load the model2vec model
    model = SentenceTransformer(model_path)
    
    print("Extracting model components...")
    
    # Create output directory
    os.makedirs(output_path, exist_ok=True)
    
    try:
        # Get the StaticEmbedding module
        static_embedding = model[0]
        
        # Extract embedding weights (move to CPU first)
        embedding_weights = static_embedding.embedding.weight.detach().cpu().numpy()
        print(f"Embedding weights shape: {embedding_weights.shape}")
        
        # Save embedding weights as numpy array
        weights_path = os.path.join(output_path, "embedding_weights.npy")
        np.save(weights_path, embedding_weights)
        print(f"Saved embedding weights to: {weights_path}")
        
        # Extract tokenizer
        tokenizer = model.tokenizer
        
        # Save tokenizer vocab
        vocab = tokenizer.get_vocab()
        vocab_path = os.path.join(output_path, "vocab.json")
        with open(vocab_path, 'w') as f:
            json.dump(vocab, f, indent=2)
        print(f"Saved vocabulary to: {vocab_path}")
        
        # Extract model config
        config_data = {
            "vocab_size": len(vocab),
            "embedding_dim": embedding_weights.shape[1],
            "model_type": "model2vec_static",
            "normalize": len(model) > 1 and hasattr(model[1], 'normalize'),
        }
        
        config_path = os.path.join(output_path, "model_config.json")
        with open(config_path, 'w') as f:
            json.dump(config_data, f, indent=2)
        print(f"Saved model config to: {config_path}")
        
        # Copy original files
        import shutil
        original_files = ['tokenizer.json', 'config.json']
        for file in original_files:
            src = os.path.join(model_path, file)
            dst = os.path.join(output_path, file)
            if os.path.exists(src):
                shutil.copy2(src, dst)
                print(f"Copied {file}")
        
        # Test the extraction
        print("Testing extraction...")
        test_extraction(model, output_path)
        
        return True
        
    except Exception as e:
        print(f"Error during extraction: {e}")
        import traceback
        traceback.print_exc()
        return False

def test_extraction(original_model, output_path):
    """Test the extracted components"""
    
    # Load extracted components
    weights = np.load(os.path.join(output_path, "embedding_weights.npy"))
    
    with open(os.path.join(output_path, "vocab.json")) as f:
        vocab = json.load(f)
    
    with open(os.path.join(output_path, "model_config.json")) as f:
        config = json.load(f)
    
    print(f"Loaded weights shape: {weights.shape}")
    print(f"Vocabulary size: {len(vocab)}")
    print(f"Config: {config}")
    
    # Test with original model
    test_text = "This is a test sentence."
    original_embedding = original_model.encode([test_text])
    
    print(f"Original embedding shape: {original_embedding.shape}")
    print(f"Original embedding (first 5 values): {original_embedding[0][:5]}")
    
    # Simple test: get token IDs and look up embeddings
    tokenizer = original_model.tokenizer
    encoding = tokenizer.encode(test_text)
    tokens = encoding.ids if hasattr(encoding, 'ids') else encoding
    print(f"Tokens: {tokens[:10]}...")  # Show first 10 tokens

    # Look up embeddings for first few tokens
    if len(tokens) > 0:
        token_embeddings = weights[tokens[:5]]  # First 5 tokens
        print(f"Token embeddings shape: {token_embeddings.shape}")
        print(f"First token embedding (first 5 values): {token_embeddings[0][:5]}")
    
    print("✅ Extraction test completed!")

def create_go_embedding_data(output_path):
    """Create Go-friendly data files"""
    
    # Load the numpy weights
    weights = np.load(os.path.join(output_path, "embedding_weights.npy"))
    
    # Convert to float32 and save as binary
    weights_f32 = weights.astype(np.float32)
    binary_path = os.path.join(output_path, "embeddings.bin")
    weights_f32.tobytes()
    
    with open(binary_path, 'wb') as f:
        f.write(weights_f32.tobytes())
    
    print(f"Saved binary embeddings to: {binary_path}")
    
    # Create metadata for Go
    metadata = {
        "vocab_size": weights.shape[0],
        "embedding_dim": weights.shape[1],
        "data_type": "float32",
        "byte_order": "little"
    }
    
    metadata_path = os.path.join(output_path, "embeddings_metadata.json")
    with open(metadata_path, 'w') as f:
        json.dump(metadata, f, indent=2)
    
    print(f"Saved metadata to: {metadata_path}")

def main():
    if len(sys.argv) != 3:
        print("Usage: python extract_embeddings.py <input_model_path> <output_path>")
        sys.exit(1)
    
    input_path = sys.argv[1]
    output_path = sys.argv[2]
    
    if not os.path.exists(input_path):
        print(f"Error: Input path does not exist: {input_path}")
        sys.exit(1)
    
    print("Model2vec Embedding Extractor")
    print("=" * 40)
    
    success = extract_model_weights(input_path, output_path)
    
    if success:
        print("\nCreating Go-friendly data files...")
        create_go_embedding_data(output_path)
        print("\n✅ Extraction completed successfully!")
        print(f"Extracted data available at: {output_path}")
    else:
        print("\n❌ Extraction failed!")
        sys.exit(1)

if __name__ == "__main__":
    main()
