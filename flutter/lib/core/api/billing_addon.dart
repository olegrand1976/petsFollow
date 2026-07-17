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

/// Active addon entitlements for the logged-in owner (cached per session load).
class AddonEntitlements {
  AddonEntitlements(this._activeCodes);

  factory AddonEntitlements.empty() => AddonEntitlements({});

  final Set<String> _activeCodes;

  bool get hasCarePlus => _activeCodes.contains('care_plus');
  bool get hasHorse => _activeCodes.contains('horse');
  bool get hasFamily => _activeCodes.contains('family');
  bool has(String code) => _activeCodes.contains(code);

  static Future<AddonEntitlements> load() async {
    try {
      final raw = await ApiClient.instance.getMyAddons();
      final codes = <String>{};
      for (final item in raw) {
        if (item is! Map) continue;
        final m = Map<String, dynamic>.from(item);
        final status = m['status'] as String? ?? '';
        final code = m['addonCode'] as String? ?? m['code'] as String? ?? '';
        if (status == 'active' && code.isNotEmpty) {
          codes.add(code);
        }
      }
      return AddonEntitlements(codes);
    } catch (_) {
      return AddonEntitlements.empty();
    }
  }
}
