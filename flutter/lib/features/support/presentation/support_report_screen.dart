import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/support/support_diagnostics_buffer.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class SupportReportScreen extends StatefulWidget {
  const SupportReportScreen({super.key, required this.source});

  /// `flutter_client` or `flutter_pro_light`
  final String source;

  @override
  State<SupportReportScreen> createState() => _SupportReportScreenState();
}

class _SupportReportScreenState extends State<SupportReportScreen> {
  final _subjectCtrl = TextEditingController();
  final _messageCtrl = TextEditingController();
  bool _saving = false;
  String? _error;
  bool _success = false;

  @override
  void dispose() {
    _subjectCtrl.dispose();
    _messageCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final l10n = AppLocalizations.of(context)!;
    final subject = _subjectCtrl.text.trim();
    final message = _messageCtrl.text.trim();
    if (subject.isEmpty || message.isEmpty || _saving) return;
    setState(() {
      _saving = true;
      _error = null;
      _success = false;
    });
    try {
      final routeName = ModalRoute.of(context)?.settings.name;
      await ApiClient.instance.createSupportTicket(
        source: widget.source,
        subject: subject,
        message: message,
        diagnostics: SupportDiagnosticsBuffer.instance.snapshot(route: routeName),
        route: routeName,
      );
      if (!mounted) return;
      setState(() => _success = true);
      await Future<void>.delayed(const Duration(milliseconds: 800));
      if (mounted) Navigator.of(context).pop(true);
    } catch (e) {
      if (!mounted) return;
      final code = apiErrorCode(e);
      setState(() {
        if (code == 'diagnostics_too_large' || code == 'payload_too_large') {
          _error = l10n.supportErrorTooLarge;
        } else {
          _error = mapApiError(e, l10n);
        }
      });
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.supportTitle)),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Text(l10n.supportHint),
          const SizedBox(height: 16),
          TextField(
            key: const Key('support_subject'),
            controller: _subjectCtrl,
            decoration: InputDecoration(labelText: l10n.supportSubject),
            maxLength: 200,
            textInputAction: TextInputAction.next,
          ),
          const SizedBox(height: 8),
          TextField(
            key: const Key('support_message'),
            controller: _messageCtrl,
            decoration: InputDecoration(labelText: l10n.supportMessage),
            maxLength: 8000,
            maxLines: 6,
          ),
          const SizedBox(height: 8),
          Text(
            l10n.supportDiagnosticsAttached,
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: AppColors.textMuted,
                ),
          ),
          if (_error != null) ...[
            const SizedBox(height: 12),
            Text(_error!, style: const TextStyle(color: AppColors.alert), key: const Key('support_error')),
          ],
          if (_success) ...[
            const SizedBox(height: 12),
            Text(l10n.supportSuccess, key: const Key('support_success')),
          ],
          const SizedBox(height: 20),
          FilledButton(
            key: const Key('support_submit'),
            onPressed: _saving ? null : _submit,
            child: Text(_saving ? l10n.supportSending : l10n.supportSubmit),
          ),
        ],
      ),
    );
  }
}

void openSupportReport(BuildContext context, {required String source}) {
  Navigator.of(context).push(
    MaterialPageRoute<void>(
      builder: (_) => SupportReportScreen(source: source),
      settings: const RouteSettings(name: '/support'),
    ),
  );
}
