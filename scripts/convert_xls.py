#!/usr/bin/env python3
"""Convert .xls files to .xlsx format"""

import sys
import os

try:
    import xlrd
    from openpyxl import Workbook
except ImportError:
    print("Please install required packages: pip install xlrd openpyxl")
    sys.exit(1)


def convert_xls_to_xlsx(xls_path, xlsx_path=None):
    """Convert .xls file to .xlsx format"""
    if xlsx_path is None:
        xlsx_path = xls_path.replace('.xls', '.xlsx')

    print(f"Converting {xls_path} to {xlsx_path}...")

    # Open the .xls file
    wb_xls = xlrd.open_workbook(xls_path)
    sheet_xls = wb_xls.sheet_by_index(0)

    # Create new .xlsx workbook
    wb_xlsx = Workbook()
    sheet_xlsx = wb_xlsx.active
    sheet_xlsx.title = sheet_xls.name[:31]  # Max 31 chars for sheet name

    # Copy data
    total_rows = sheet_xls.nrows
    for row_idx in range(total_rows):
        for col_idx in range(sheet_xls.ncols):
            value = sheet_xls.cell_value(row_idx, col_idx)
            sheet_xlsx.cell(row=row_idx + 1, column=col_idx + 1, value=value)

        if (row_idx + 1) % 10000 == 0:
            print(f"  Processed {row_idx + 1}/{total_rows} rows...")

    # Save
    wb_xlsx.save(xlsx_path)
    print(f"Converted {total_rows} rows to {xlsx_path}")
    return xlsx_path


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python convert_xls.py file1.xls [file2.xls ...]")
        sys.exit(1)

    for xls_file in sys.argv[1:]:
        if os.path.exists(xls_file):
            convert_xls_to_xlsx(xls_file)
        else:
            print(f"File not found: {xls_file}")
