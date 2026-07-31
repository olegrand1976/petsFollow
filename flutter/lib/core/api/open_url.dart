import 'package:url_launcher/url_launcher.dart';

/// Test seam — override in widget tests to avoid hanging on [launchUrl].
typedef OpenExternalUrlFn = Future<bool> Function(String url);

OpenExternalUrlFn openExternalUrlImpl = _defaultOpenExternalUrl;

Future<bool> _defaultOpenExternalUrl(String url) async {
  final uri = Uri.tryParse(url);
  if (uri == null) return false;
  try {
    // Prefer in-app browser / Custom Tabs to reduce process death on return.
    if (await launchUrl(uri, mode: LaunchMode.inAppBrowserView)) {
      return true;
    }
    return await launchUrl(uri, mode: LaunchMode.externalApplication);
  } catch (_) {
    try {
      return await launchUrl(uri, mode: LaunchMode.externalApplication);
    } catch (_) {
      return false;
    }
  }
}

/// Opens [url] externally. Does not gate on [canLaunchUrl] (unreliable on Android 11+).
Future<bool> openExternalUrl(String url) => openExternalUrlImpl(url);
