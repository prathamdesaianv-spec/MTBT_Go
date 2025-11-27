#!/usr/bin/env python3
"""
PDF to Text Converter
Converts PDF files to text files using PyPDF2
"""

import os
from PyPDF2 import PdfReader

def pdf_to_text(pdf_path, output_path):
    """
    Convert a PDF file to text file
    
    Args:
        pdf_path: Path to the input PDF file
        output_path: Path to the output text file
    """
    try:
        # Read the PDF file
        reader = PdfReader(pdf_path)
        
        # Extract text from all pages
        text_content = []
        print(f"Processing {os.path.basename(pdf_path)}...")
        print(f"Total pages: {len(reader.pages)}")
        
        for page_num, page in enumerate(reader.pages, 1):
            text = page.extract_text()
            text_content.append(f"--- Page {page_num} ---\n{text}\n")
            print(f"Extracted page {page_num}/{len(reader.pages)}")
        
        # Write to text file
        with open(output_path, 'w', encoding='utf-8') as f:
            f.write('\n'.join(text_content))
        
        print(f"✓ Successfully converted to: {output_path}\n")
        return True
        
    except Exception as e:
        print(f"✗ Error processing {pdf_path}: {str(e)}\n")
        return False

def main():
    # Define the PDF files to convert
    base_dir = "/teamspace/studios/this_studio/MTBT_Go"
    pdf_files = [
        f"{base_dir}/NSE docs/MSD67677.pdf",
        f"{base_dir}/NSE docs/MTBT_FO_NNF_PROTOCL_6.7_0 (1).pdf"
    ]
    
    # Output directory for text files
    output_dir = f"{base_dir}/text_convert/output"
    os.makedirs(output_dir, exist_ok=True)
    
    print("=" * 60)
    print("PDF to Text Converter")
    print("=" * 60 + "\n")
    
    # Convert each PDF
    success_count = 0
    for pdf_path in pdf_files:
        if not os.path.exists(pdf_path):
            print(f"✗ File not found: {pdf_path}\n")
            continue
        
        # Create output filename
        pdf_filename = os.path.basename(pdf_path)
        text_filename = os.path.splitext(pdf_filename)[0] + ".txt"
        output_path = os.path.join(output_dir, text_filename)
        
        # Convert
        if pdf_to_text(pdf_path, output_path):
            success_count += 1
    
    print("=" * 60)
    print(f"Conversion complete: {success_count}/{len(pdf_files)} files converted")
    print(f"Output location: {output_dir}")
    print("=" * 60)

if __name__ == "__main__":
    main()
