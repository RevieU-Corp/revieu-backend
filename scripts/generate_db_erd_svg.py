#!/usr/bin/env python3
"""
Generate a full PostgreSQL ERD SVG with PK/FK markers.

Example:
  python scripts/generate_db_erd_svg.py \
    --host 10.0.0.1 --user postgres --db revieu --password 123456 \
    --output docs/database-erd.svg
"""

from __future__ import annotations

import argparse
import html
import os
import subprocess
import sys
from collections import defaultdict
from dataclasses import dataclass


@dataclass
class Column:
    name: str
    data_type: str
    not_null: bool


@dataclass
class FK:
    constraint: str
    table: str
    column: str
    ref_table: str
    ref_column: str


def run_psql(host: str, user: str, db: str, password: str, sql: str) -> list[str]:
    env = os.environ.copy()
    env["PGPASSWORD"] = password
    cmd = [
        "psql",
        "-h",
        host,
        "-U",
        user,
        "-d",
        db,
        "-At",
        "-F",
        "\t",
        "-c",
        sql,
    ]
    result = subprocess.run(
        cmd,
        env=env,
        check=False,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        raise RuntimeError(result.stderr.strip() or "psql failed")
    return [line for line in result.stdout.splitlines() if line.strip()]


def load_schema(host: str, user: str, db: str, password: str):
    tables = run_psql(
        host,
        user,
        db,
        password,
        """
        SELECT c.relname
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public' AND c.relkind = 'r'
        ORDER BY c.relname;
        """,
    )

    col_rows = run_psql(
        host,
        user,
        db,
        password,
        """
        SELECT c.relname,
               a.attnum,
               a.attname,
               pg_catalog.format_type(a.atttypid, a.atttypmod) AS data_type,
               a.attnotnull::text
        FROM pg_attribute a
        JOIN pg_class c ON c.oid = a.attrelid
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public'
          AND c.relkind = 'r'
          AND a.attnum > 0
          AND NOT a.attisdropped
        ORDER BY c.relname, a.attnum;
        """,
    )

    pk_rows = run_psql(
        host,
        user,
        db,
        password,
        """
        SELECT c.conrelid::regclass::text AS table_name,
               a.attname AS column_name
        FROM pg_constraint c
        JOIN unnest(c.conkey) WITH ORDINALITY AS k(attnum, ord) ON TRUE
        JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = k.attnum
        JOIN pg_namespace n ON n.oid = c.connamespace
        WHERE c.contype = 'p' AND n.nspname = 'public'
        ORDER BY table_name, k.ord;
        """,
    )

    fk_rows = run_psql(
        host,
        user,
        db,
        password,
        """
        SELECT con.conname,
               con.conrelid::regclass::text AS table_name,
               src.attname AS column_name,
               con.confrelid::regclass::text AS ref_table,
               tgt.attname AS ref_column
        FROM pg_constraint con
        JOIN pg_namespace n ON n.oid = con.connamespace
        JOIN generate_subscripts(con.conkey, 1) AS s(i) ON TRUE
        JOIN pg_attribute src ON src.attrelid = con.conrelid AND src.attnum = con.conkey[s.i]
        JOIN pg_attribute tgt ON tgt.attrelid = con.confrelid AND tgt.attnum = con.confkey[s.i]
        WHERE con.contype = 'f' AND n.nspname = 'public'
        ORDER BY table_name, con.conname, s.i;
        """,
    )

    columns: dict[str, list[Column]] = defaultdict(list)
    for row in col_rows:
        table, _attnum, name, data_type, notnull_txt = row.split("\t")
        columns[table].append(
            Column(name=name, data_type=data_type, not_null=(notnull_txt == "t"))
        )

    pk_cols: set[tuple[str, str]] = set()
    for row in pk_rows:
        table, column = row.split("\t")
        pk_cols.add((table, column))

    fks: list[FK] = []
    fks_by_col: dict[tuple[str, str], list[FK]] = defaultdict(list)
    for row in fk_rows:
        conname, table, column, ref_table, ref_column = row.split("\t")
        fk = FK(
            constraint=conname,
            table=table,
            column=column,
            ref_table=ref_table,
            ref_column=ref_column,
        )
        fks.append(fk)
        fks_by_col[(table, column)].append(fk)

    return tables, columns, pk_cols, fks, fks_by_col


def make_svg(
    tables: list[str],
    columns: dict[str, list[Column]],
    pk_cols: set[tuple[str, str]],
    fks: list[FK],
    fks_by_col: dict[tuple[str, str], list[FK]],
) -> str:
    ncols = 6
    margin_x = 40
    margin_y = 140
    col_width = 430
    row_gap = 32
    line_h = 16
    header_h = 28
    node_pad = 10

    node_meta = {}
    row_heights: list[int] = []

    for idx, table in enumerate(tables):
        lines = []
        col_index = {}
        max_chars = len(table) + 4
        for i, col in enumerate(columns.get(table, [])):
            col_index[col.name] = i
            markers = []
            if (table, col.name) in pk_cols:
                markers.append("PK")
            if (table, col.name) in fks_by_col:
                markers.append("FK")
            marker_txt = f" [{','.join(markers)}]" if markers else ""
            fk_targets = ""
            if (table, col.name) in fks_by_col:
                refs = ", ".join(
                    f"{fk.ref_table}.{fk.ref_column}"
                    for fk in fks_by_col[(table, col.name)]
                )
                fk_targets = f" -> {refs}"
            null_txt = "" if col.not_null else " ?"
            line = f"{col.name}: {col.data_type}{null_txt}{marker_txt}{fk_targets}"
            lines.append(line)
            max_chars = max(max_chars, len(line))

        width = min(max(250, int(max_chars * 6.8) + 24), 410)
        height = header_h + len(lines) * line_h + node_pad * 2
        row = idx // ncols
        while len(row_heights) <= row:
            row_heights.append(0)
        row_heights[row] = max(row_heights[row], height)
        node_meta[table] = {
            "idx": idx,
            "row": row,
            "col": idx % ncols,
            "width": width,
            "height": height,
            "lines": lines,
            "col_index": col_index,
        }

    row_offsets = []
    y = margin_y
    for h in row_heights:
        row_offsets.append(y)
        y += h + row_gap

    total_width = margin_x * 2 + ncols * col_width
    total_height = y + 40

    for table in tables:
        meta = node_meta[table]
        col = meta["col"]
        row = meta["row"]
        w = meta["width"]
        h = meta["height"]
        x = margin_x + col * col_width + (col_width - w) // 2
        y0 = row_offsets[row] + (row_heights[row] - h) // 2
        meta["x"] = x
        meta["y"] = y0

    def anchor(table: str, column: str, side_hint: str):
        meta = node_meta[table]
        x = meta["x"]
        y0 = meta["y"]
        w = meta["width"]
        idx = meta["col_index"].get(column, 0)
        y = y0 + header_h + node_pad + int((idx + 0.6) * line_h)
        if side_hint == "left":
            return x, y
        if side_hint == "right":
            return x + w, y
        return x + w // 2, y

    svg_parts = []
    svg_parts.append(
        f'<svg xmlns="http://www.w3.org/2000/svg" width="{total_width}" '
        f'height="{total_height}" viewBox="0 0 {total_width} {total_height}">'
    )
    svg_parts.append(
        "<defs>"
        '<marker id="arrow" markerWidth="10" markerHeight="10" refX="9" refY="3" '
        'orient="auto" markerUnits="strokeWidth">'
        '<path d="M0,0 L0,6 L9,3 z" fill="#64748b" />'
        "</marker>"
        "</defs>"
    )
    svg_parts.append(
        f'<rect x="0" y="0" width="{total_width}" height="{total_height}" fill="#f8fafc" />'
    )

    title = "RevieU Database ERD (public schema)"
    subtitle = f"Tables: {len(tables)} | FK relationships: {len(fks)} | PK/FK markers shown per column"
    svg_parts.append(
        f'<text x="{margin_x}" y="40" font-family="Arial, sans-serif" font-size="24" '
        f'font-weight="700" fill="#0f172a">{html.escape(title)}</text>'
    )
    svg_parts.append(
        f'<text x="{margin_x}" y="66" font-family="Arial, sans-serif" font-size="13" '
        f'fill="#334155">{html.escape(subtitle)}</text>'
    )
    svg_parts.append(
        f'<text x="{margin_x}" y="88" font-family="Arial, sans-serif" font-size="12" '
        f'fill="#334155">Legend: [PK] primary key, [FK] foreign key, "?" nullable column</text>'
    )

    # Draw edges first.
    for fk in fks:
        s_meta = node_meta[fk.table]
        t_meta = node_meta[fk.ref_table]
        s_cx = s_meta["x"] + s_meta["width"] / 2
        t_cx = t_meta["x"] + t_meta["width"] / 2

        if s_cx < t_cx:
            sx, sy = anchor(fk.table, fk.column, "right")
            tx, ty = anchor(fk.ref_table, fk.ref_column, "left")
        elif s_cx > t_cx:
            sx, sy = anchor(fk.table, fk.column, "left")
            tx, ty = anchor(fk.ref_table, fk.ref_column, "right")
        else:
            s_top = s_meta["y"]
            t_top = t_meta["y"]
            if s_top < t_top:
                sx, sy = anchor(fk.table, fk.column, "center")
                tx, ty = anchor(fk.ref_table, fk.ref_column, "center")
                sy = s_meta["y"] + s_meta["height"]
                ty = t_meta["y"]
            else:
                sx, sy = anchor(fk.table, fk.column, "center")
                tx, ty = anchor(fk.ref_table, fk.ref_column, "center")
                sy = s_meta["y"]
                ty = t_meta["y"] + t_meta["height"]

        if abs(sx - tx) >= abs(sy - ty):
            mx = (sx + tx) / 2
            path = f"M {sx:.1f} {sy:.1f} C {mx:.1f} {sy:.1f}, {mx:.1f} {ty:.1f}, {tx:.1f} {ty:.1f}"
        else:
            my = (sy + ty) / 2
            path = f"M {sx:.1f} {sy:.1f} C {sx:.1f} {my:.1f}, {tx:.1f} {my:.1f}, {tx:.1f} {ty:.1f}"

        edge_title = f"{fk.table}.{fk.column} -> {fk.ref_table}.{fk.ref_column} ({fk.constraint})"
        svg_parts.append(
            f'<path d="{path}" stroke="#64748b" stroke-opacity="0.45" stroke-width="1.1" '
            f'fill="none" marker-end="url(#arrow)"><title>{html.escape(edge_title)}</title></path>'
        )

    # Draw table nodes.
    for table in tables:
        meta = node_meta[table]
        x = meta["x"]
        y0 = meta["y"]
        w = meta["width"]
        h = meta["height"]
        lines = meta["lines"]

        svg_parts.append(
            f'<rect x="{x}" y="{y0}" width="{w}" height="{h}" rx="8" ry="8" '
            'fill="#ffffff" stroke="#0f172a" stroke-opacity="0.25" />'
        )
        svg_parts.append(
            f'<rect x="{x}" y="{y0}" width="{w}" height="{header_h + 4}" rx="8" ry="8" '
            'fill="#e2e8f0" stroke="none" />'
        )
        svg_parts.append(
            f'<text x="{x + 10}" y="{y0 + 20}" font-family="Arial, sans-serif" '
            f'font-size="14" font-weight="700" fill="#0f172a">{html.escape(table)}</text>'
        )
        svg_parts.append(
            f'<line x1="{x}" y1="{y0 + header_h + 4}" x2="{x + w}" y2="{y0 + header_h + 4}" '
            'stroke="#94a3b8" stroke-width="1" />'
        )

        for i, line in enumerate(lines):
            ty = y0 + header_h + node_pad + (i + 1) * line_h
            svg_parts.append(
                f'<text x="{x + 10}" y="{ty}" font-family="Courier New, monospace" '
                f'font-size="11" fill="#0f172a">{html.escape(line)}</text>'
            )

    svg_parts.append("</svg>")
    return "\n".join(svg_parts)


def main() -> int:
    parser = argparse.ArgumentParser(description="Generate PostgreSQL ERD SVG")
    parser.add_argument("--host", required=True, help="PostgreSQL host")
    parser.add_argument("--user", required=True, help="PostgreSQL user")
    parser.add_argument("--db", required=True, help="PostgreSQL database")
    parser.add_argument("--password", required=True, help="PostgreSQL password")
    parser.add_argument("--output", default="docs/database-erd.svg", help="Output SVG path")
    args = parser.parse_args()

    tables, columns, pk_cols, fks, fks_by_col = load_schema(
        host=args.host,
        user=args.user,
        db=args.db,
        password=args.password,
    )
    svg = make_svg(
        tables=tables,
        columns=columns,
        pk_cols=pk_cols,
        fks=fks,
        fks_by_col=fks_by_col,
    )
    out_path = args.output
    out_dir = os.path.dirname(out_path) or "."
    os.makedirs(out_dir, exist_ok=True)
    with open(out_path, "w", encoding="utf-8") as f:
        f.write(svg)

    print(f"Generated: {out_path}")
    print(f"Tables: {len(tables)}")
    print(f"FK columns: {len(fks)}")
    return 0


if __name__ == "__main__":
    sys.exit(main())

