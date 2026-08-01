#!/usr/bin/env python3
"""Generate a minimal Explicit VR Little Endian DX-style DICOM for Orthanc demo.

Includes PixelSpacing / ImagerPixelSpacing so the Pro canvas viewer can measure in mm.
"""
from __future__ import annotations

import argparse
import struct
from pathlib import Path

# Known isotropic spacing used by demos, e2e and Vitest expectations (mm per pixel).
DEMO_PIXEL_SPACING_MM = 0.5


def _even(b: bytes) -> bytes:
    return b if len(b) % 2 == 0 else b + b"\x00"


def _tag(group: int, element: int, vr: str, value: bytes) -> bytes:
    """Encode one DICOM data element (explicit VR LE)."""
    if vr in ("OB", "OW", "OF", "SQ", "UT", "UN"):
        value = _even(value) if vr == "OB" else value
        return struct.pack("<HH2sH", group, element, vr.encode("ascii"), 0) + struct.pack(
            "<I", len(value)
        ) + value
    value = _even(value)
    return struct.pack("<HH2sH", group, element, vr.encode("ascii"), len(value)) + value


def _ascii(s: str) -> bytes:
    return _even(s.encode("ascii"))


def _ds_pair(row_mm: float, col_mm: float) -> bytes:
    """DICOM DS multi-value: row\\col (mm)."""
    return _ascii(f"{row_mm}\\{col_mm}")


def build_dicom(
    *,
    patient_name: str = "Rex^Demo",
    patient_id: str = "PF-DEMO-REX",
    study_desc: str = "RX thorax demo petsFollow",
    rows: int = 64,
    cols: int = 64,
    spacing_mm: float = DEMO_PIXEL_SPACING_MM,
) -> bytes:
    # Tiny gradient image (8-bit mono).
    pixels = _even(bytes((i * 4) % 256 for i in range(rows * cols)))

    sop_class = "1.2.840.10008.5.1.4.1.1.7"  # Secondary Capture (widely accepted by Orthanc)
    # Bump trailing segment when tags change — Orthanc OverwriteInstances=false keeps old blobs.
    sop_instance = "1.2.826.0.1.3680043.8.498.1001.3"
    study_uid = "1.2.826.0.1.3680043.8.498.1001.10"
    series_uid = "1.2.826.0.1.3680043.8.498.1001.11"
    transfer_syntax = "1.2.840.10008.1.2.1"  # Explicit VR Little Endian
    spacing = _ds_pair(spacing_mm, spacing_mm)

    meta = b"".join(
        [
            _tag(0x0002, 0x0001, "OB", b"\x00\x01"),
            _tag(0x0002, 0x0002, "UI", _ascii(sop_class)),
            _tag(0x0002, 0x0003, "UI", _ascii(sop_instance)),
            _tag(0x0002, 0x0010, "UI", _ascii(transfer_syntax)),
            _tag(0x0002, 0x0012, "UI", _ascii("1.2.826.0.1.3680043.8.498.1")),
            _tag(0x0002, 0x0013, "SH", _ascii("petsFollow")),
        ]
    )
    meta_with_len = _tag(0x0002, 0x0000, "UL", struct.pack("<I", len(meta))) + meta

    ds = b"".join(
        [
            _tag(0x0008, 0x0016, "UI", _ascii(sop_class)),
            _tag(0x0008, 0x0018, "UI", _ascii(sop_instance)),
            _tag(0x0008, 0x0020, "DA", _ascii("20260101")),
            _tag(0x0008, 0x0030, "TM", _ascii("120000")),
            _tag(0x0008, 0x0060, "CS", _ascii("DX")),
            _tag(0x0008, 0x1030, "LO", _ascii(study_desc)),
            _tag(0x0010, 0x0010, "PN", _ascii(patient_name)),
            _tag(0x0010, 0x0020, "LO", _ascii(patient_id)),
            _tag(0x0018, 0x1164, "DS", spacing),  # ImagerPixelSpacing
            _tag(0x0020, 0x000D, "UI", _ascii(study_uid)),
            _tag(0x0020, 0x000E, "UI", _ascii(series_uid)),
            _tag(0x0020, 0x0011, "IS", _ascii("1")),
            _tag(0x0020, 0x0013, "IS", _ascii("1")),
            _tag(0x0028, 0x0002, "US", struct.pack("<H", 1)),
            _tag(0x0028, 0x0004, "CS", _ascii("MONOCHROME2")),
            _tag(0x0028, 0x0010, "US", struct.pack("<H", rows)),
            _tag(0x0028, 0x0011, "US", struct.pack("<H", cols)),
            _tag(0x0028, 0x0030, "DS", spacing),  # PixelSpacing
            _tag(0x0028, 0x0100, "US", struct.pack("<H", 8)),
            _tag(0x0028, 0x0101, "US", struct.pack("<H", 8)),
            _tag(0x0028, 0x0102, "US", struct.pack("<H", 7)),
            _tag(0x0028, 0x0103, "US", struct.pack("<H", 0)),
            _tag(0x7FE0, 0x0010, "OB", pixels),
        ]
    )

    return b"\x00" * 128 + b"DICM" + meta_with_len + ds


def main() -> None:
    p = argparse.ArgumentParser()
    p.add_argument(
        "-o",
        "--output",
        type=Path,
        default=Path(__file__).resolve().parents[1] / "testdata" / "pacs" / "demo-rx.dcm",
    )
    args = p.parse_args()
    args.output.parent.mkdir(parents=True, exist_ok=True)
    data = build_dicom()
    args.output.write_bytes(data)
    print(
        f"wrote {args.output} ({len(data)} bytes) "
        f"PixelSpacing={DEMO_PIXEL_SPACING_MM}\\{DEMO_PIXEL_SPACING_MM} mm"
    )


if __name__ == "__main__":
    main()
