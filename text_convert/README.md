# PDF to Text Converter

This folder contains a Python script to convert PDF files to text format using PyPDF2.

## Installation

Install the required library:
```bash
pip install PyPDF2
```

## Usage

Run the conversion script:
```bash
python pdf_to_text.py
```

The script will:
1. Convert the specified PDF files from the `NSE docs/` folder
2. Save the extracted text files in the `text_convert/output/` directory
3. Display progress and completion status

## Output

Text files will be saved in:
- `text_convert/output/MSD67677.txt`
- `text_convert/output/MTBT_FO_NNF_PROTOCL_6.7_0 (1).txt`

Each text file will contain the extracted text with page numbers marked for reference.
