import 'package:petsfollow_mobile/core/api/api_client.dart';

/// Rewrites `localhost` / `127.0.0.1` media URLs to the configured API host
/// so emulators and physical devices can load `/media/...` assets.
String? resolveMediaUrl(String? url) {
  if (url == null || url.isEmpty) return url;
  final uri = Uri.tryParse(url);
  if (uri == null || !uri.hasScheme || uri.host.isEmpty) return url;
  final host = uri.host.toLowerCase();
  if (host != 'localhost' && host != '127.0.0.1') return url;
  final base = Uri.tryParse(ApiClient.instance.dio.options.baseUrl);
  if (base == null || base.host.isEmpty) return url;
  return uri
      .replace(
        scheme: base.scheme.isNotEmpty ? base.scheme : uri.scheme,
        host: base.host,
        port: base.hasPort ? base.port : null,
      )
      .toString();
}
