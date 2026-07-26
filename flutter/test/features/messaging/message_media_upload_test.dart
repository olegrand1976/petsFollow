import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/features/messaging/message_media_upload.dart';

void main() {
  test('keeps original extension when not compressed', () {
    expect(
      messageMediaUploadBasename('/tmp/clip.MOV', fromCompressedOutput: false),
      'clip.MOV',
    );
    expect(
      messageMediaUploadBasename('/tmp/clip.webm', fromCompressedOutput: false),
      'clip.webm',
    );
  });

  test('rewrites to .mp4 only for compressed output', () {
    expect(
      messageMediaUploadBasename('/tmp/clip.MOV', fromCompressedOutput: true),
      'clip.mp4',
    );
    expect(
      messageMediaUploadBasename('/cache/out.mp4', fromCompressedOutput: true),
      'out.mp4',
    );
  });

  test('compress threshold', () {
    expect(shouldCompressMessageVideo(kVideoCompressThresholdBytes), isFalse);
    expect(shouldCompressMessageVideo(kVideoCompressThresholdBytes + 1), isTrue);
    expect(shouldCompressMessageVideo(kMaxMessageMediaBytes), isTrue);
  });
}
