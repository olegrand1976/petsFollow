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
