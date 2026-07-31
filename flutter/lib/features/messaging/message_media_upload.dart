import 'dart:io';

/// Aligné sur `media.MaxMessageMediaBytes` (Go / GCS).
const kMaxMessageMediaBytes = 25 << 20;

/// Au-delà, compression locale avant upload.
const kVideoCompressThresholdBytes = 8 << 20;

const kMaxMessageVideoDuration = Duration(minutes: 2);

/// Nom multipart : conserver l’extension réelle sauf sortie `video_compress` (MP4).
String messageMediaUploadBasename(
  String path, {
  required bool fromCompressedOutput,
}) {
  final base = path.split(Platform.pathSeparator).last;
  if (!fromCompressedOutput) return base;
  if (base.toLowerCase().endsWith('.mp4')) return base;
  final dot = base.lastIndexOf('.');
  final stem = dot > 0 ? base.substring(0, dot) : base;
  return '$stem.mp4';
}

bool shouldCompressMessageVideo(int sizeBytes) =>
    sizeBytes > kVideoCompressThresholdBytes;
