import 'package:petsfollow_mobile/core/api/api_client.dart';

class BillingAddon {
  const BillingAddon({
    required this.code,
    required this.label,
    required this.amountCents,
    required this.currency,
  });

  final String code;
  final String label;
  final int amountCents;
  final String currency;

  factory BillingAddon.fromJson(Map<String, dynamic> json) {
    return BillingAddon(
      code: json['code'] as String? ?? '',
      label: json['label'] as String? ?? '',
      amountCents: (json['amountCents'] as num?)?.toInt() ?? 0,
      currency: json['currency'] as String? ?? 'eur',
    );
  }

  static Future<List<BillingAddon>> fetchCatalog() async {
    final raw = await ApiClient.instance.getBillingAddons();
    return raw
        .whereType<Map>()
        .map((e) => BillingAddon.fromJson(Map<String, dynamic>.from(e)))
        .where((a) => a.code.isNotEmpty)
        .toList();
  }

  static BillingAddon? byCode(List<BillingAddon> catalog, String code) {
    for (final a in catalog) {
      if (a.code == code) return a;
    }
    return null;
  }
}

/// Former paid addons (Care+/Horse/Family/Kennel) are unlocked for all owners.
///
/// Feature getters always return `true` so UI never gates on addon purchase.
/// [load] still probes `/billing/my-addons` for callers that need API health.
class AddonEntitlements {
  const AddonEntitlements();

  factory AddonEntitlements.empty() => const AddonEntitlements();

  bool get hasCarePlus => true;
  bool get hasHorse => true;
  bool get hasFamily => true;
  bool get hasKennel => true;
  bool has(String code) => true;

  /// Returns entitlements, or `null` if the request failed.
  static Future<AddonEntitlements?> load() async {
    try {
      await ApiClient.instance.getMyAddons();
      return const AddonEntitlements();
    } catch (_) {
      return null;
    }
  }
}
