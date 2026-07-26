import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/auth/google_auth.dart';
import 'package:petsfollow_mobile/core/auth/google_login_flow.dart';
import 'package:petsfollow_mobile/core/invite/invite_code_store.dart';
import 'package:petsfollow_mobile/core/locale/locale_controller.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/app_theme.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/core/widgets/pets_logo.dart';
import 'package:petsfollow_mobile/features/legal/domain/legal_document_type.dart';
import 'package:petsfollow_mobile/features/legal/presentation/legal_document_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class RegisterScreen extends StatefulWidget {
  const RegisterScreen({super.key});

  @override
  State<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends State<RegisterScreen> {
  final fullName = TextEditingController();
  final email = TextEditingController();
  final password = TextEditingController();
  final confirm = TextEditingController();
  final postalCode = TextEditingController();
  final inviteCodeCtrl = TextEditingController();
  String? error;
  String? info;
  String? success;
  bool _busy = false;
  bool consent = false;
  bool _hasInvite = false;
  bool _nearbyLoading = false;
  bool _nearbyEmpty = false;
  String? _nearbyError;
  String? _selectedCommercialId;
  List<Map<String, dynamic>> _nearby = [];

  bool get _isIOS => defaultTargetPlatform == TargetPlatform.iOS;

  String get _effectiveInviteCode => inviteCodeCtrl.text.trim().toUpperCase();

  bool get _hasInviteCode => _effectiveInviteCode.isNotEmpty;

  @override
  void initState() {
    super.initState();
    InviteCodeStore.instance.peek().then((code) {
      if (!mounted) return;
      if (code != null && code.isNotEmpty) {
        inviteCodeCtrl.text = code;
      }
      setState(() => _hasInvite = code != null && code.isNotEmpty);
    });
    inviteCodeCtrl.addListener(() {
      final has = _hasInviteCode;
      if (has != _hasInvite) {
        setState(() {
          _hasInvite = has;
          if (has) _selectedCommercialId = null;
        });
      }
    });
  }

  @override
  void dispose() {
    fullName.dispose();
    email.dispose();
    password.dispose();
    confirm.dispose();
    postalCode.dispose();
    inviteCodeCtrl.dispose();
    super.dispose();
  }

  Future<void> submit() async {
    final l10n = AppLocalizations.of(context)!;
    final name = fullName.text.trim();
    final mail = email.text.trim();
    final pass = password.text;
    if (name.isEmpty || mail.isEmpty || !mail.contains('@')) {
      setState(() => error = l10n.emailRequired);
      return;
    }
    if (pass.length < 8) {
      setState(() => error = l10n.passwordTooShort);
      return;
    }
    if (pass != confirm.text) {
      setState(() => error = l10n.passwordMismatch);
      return;
    }
    if (!consent) {
      setState(() => error = l10n.registerConsentRequired);
      return;
    }
    setState(() {
      error = null;
      info = null;
      success = null;
      _busy = true;
    });
    try {
      final result = await ApiClient.instance.registerClient(
        email: mail,
        password: pass,
        fullName: name,
        locale: LocaleController.instance.locale.languageCode,
        consent: consent,
        inviteCode: _hasInviteCode ? _effectiveInviteCode : null,
        commercialUserId: _hasInviteCode ? null : _selectedCommercialId,
      );
      if (!mounted) return;
      final inviteStatus = result['inviteStatus']?.toString() ?? '';
      final inviteOk = inviteStatus == 'referred' ||
          inviteStatus == 'linked' ||
          inviteStatus == 'granted' ||
          inviteStatus == 'already_linked';
      setState(() {
        success = _hasInviteCode && !inviteOk
            ? l10n.registerInviteNotApplied
            : l10n.registerSuccess;
      });
    } on DioException catch (e) {
      if (!mounted) return;
      final code = apiErrorCode(e);
      setState(() {
        error = code == 'email_already_exists' || code == 'conflict'
            ? l10n.registerEmailExists
            : l10n.registerFailed;
      });
    } catch (_) {
      if (mounted) setState(() => error = l10n.registerFailed);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> submitGoogle() async {
    final l10n = AppLocalizations.of(context)!;
    if (!GoogleAuth.isConfigured) {
      setState(() => error = l10n.googleNotConfigured);
      return;
    }
    if (!consent) {
      setState(() => error = l10n.registerConsentRequired);
      return;
    }
    setState(() {
      error = null;
      info = null;
      _busy = true;
    });
    try {
      if (_hasInviteCode) {
        await InviteCodeStore.instance.save(_effectiveInviteCode);
      }
      final data = await GoogleLoginFlow.signIn(
        consent: true,
        commercialUserId: _hasInviteCode ? null : _selectedCommercialId,
      );
      if (!mounted) return;
      if (data != null) {
        // L'utilisateur est connecté (ou en attente de 2FA) : le LoginScreen
        // parent finalise via _finishLogin.
        Navigator.of(context).pop(data);
        return;
      }
      setState(() => _busy = false);
    } catch (e) {
      if (!mounted) return;
      setState(() {
        error = GoogleLoginFlow.errorMessage(l10n, e);
        _busy = false;
      });
    }
  }

  void _openLegal(LegalDocumentType type) {
    Navigator.of(context).push(
      MaterialPageRoute<void>(builder: (_) => LegalDocumentScreen(type: type)),
    );
  }

  Widget _buildConsentRow(AppLocalizations l10n) {
    final linkStyle = TextStyle(
      color: AppColors.accent,
      decoration: TextDecoration.underline,
      decorationColor: AppColors.accent,
    );
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Checkbox(
          value: consent,
          onChanged: _busy ? null : (v) => setState(() => consent = v ?? false),
        ),
        Expanded(
          child: Padding(
            padding: const EdgeInsets.only(top: 12),
            child: Wrap(
              children: [
                Text(l10n.registerConsentPrefix),
                GestureDetector(
                  onTap: () => _openLegal(LegalDocumentType.terms),
                  child: Text(l10n.legalTermsTitle, style: linkStyle),
                ),
                Text(l10n.registerConsentMiddle),
                GestureDetector(
                  onTap: () => _openLegal(LegalDocumentType.privacy),
                  child: Text(l10n.legalPrivacyTitle, style: linkStyle),
                ),
                const Text('.'),
              ],
            ),
          ),
        ),
      ],
    );
  }

  void tapApple() {
    final l10n = AppLocalizations.of(context)!;
    setState(() {
      error = null;
      info = l10n.appleComingSoon;
    });
  }

  Future<void> _findByGeo() async {
    final l10n = AppLocalizations.of(context)!;
    setState(() {
      _nearbyError = null;
      _nearbyLoading = true;
      _nearbyEmpty = false;
    });
    try {
      var permission = await Geolocator.checkPermission();
      if (permission == LocationPermission.denied) {
        permission = await Geolocator.requestPermission();
      }
      if (permission == LocationPermission.denied ||
          permission == LocationPermission.deniedForever) {
        setState(() {
          _nearbyLoading = false;
          _nearbyError = l10n.nearbyCommercialGeoDenied;
        });
        return;
      }
      final pos = await Geolocator.getCurrentPosition();
      final rows = await ApiClient.instance.listNearbyCommercials(
        lat: pos.latitude,
        lng: pos.longitude,
      );
      if (!mounted) return;
      setState(() {
        _nearby = rows;
        _nearbyEmpty = rows.isEmpty;
        _nearbyLoading = false;
        if (!_nearby.any((r) => r['userId'] == _selectedCommercialId)) {
          _selectedCommercialId = null;
        }
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _nearbyLoading = false;
        _nearbyError = l10n.nearbyCommercialGeoDenied;
      });
    }
  }

  Future<void> _findByPostal() async {
    final code = postalCode.text.trim();
    if (code.isEmpty) return;
    setState(() {
      _nearbyError = null;
      _nearbyLoading = true;
      _nearbyEmpty = false;
    });
    try {
      final rows = await ApiClient.instance.listNearbyCommercials(postalCode: code);
      if (!mounted) return;
      setState(() {
        _nearby = rows;
        _nearbyEmpty = rows.isEmpty;
        _nearbyLoading = false;
        if (!_nearby.any((r) => r['userId'] == _selectedCommercialId)) {
          _selectedCommercialId = null;
        }
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _nearbyLoading = false;
        _nearbyError = AppLocalizations.of(context)!.registerFailed;
      });
    }
  }

  Widget _buildNearbySection(AppLocalizations l10n) {
    if (_hasInviteCode) return const SizedBox.shrink();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const SizedBox(height: 16),
        Text(l10n.nearbyCommercialTitle, style: Theme.of(context).textTheme.titleSmall),
        const SizedBox(height: 4),
        Text(l10n.nearbyCommercialHint, style: Theme.of(context).textTheme.bodySmall),
        const SizedBox(height: 8),
        OutlinedButton(
          onPressed: (_busy || _nearbyLoading) ? null : _findByGeo,
          child: Text(l10n.nearbyCommercialUseLocation),
        ),
        const SizedBox(height: 8),
        Row(
          children: [
            Expanded(
              child: TextField(
                controller: postalCode,
                keyboardType: TextInputType.number,
                decoration: InputDecoration(labelText: l10n.nearbyCommercialPostalCode),
              ),
            ),
            const SizedBox(width: 8),
            FilledButton(
              onPressed: (_busy || _nearbyLoading) ? null : _findByPostal,
              child: Text(l10n.nearbyCommercialSearch),
            ),
          ],
        ),
        if (_nearbyLoading) ...[
          const SizedBox(height: 8),
          const Center(child: CircularProgressIndicator()),
        ],
        if (_nearbyError != null) ...[
          const SizedBox(height: 8),
          Text(_nearbyError!, style: const TextStyle(color: AppColors.alert)),
        ],
        if (_nearbyEmpty) ...[
          const SizedBox(height: 8),
          Text(l10n.nearbyCommercialEmpty),
        ],
        ..._nearby.map((row) {
          final id = row['userId']?.toString() ?? '';
          final name = row['fullName']?.toString() ?? '';
          final city = row['city']?.toString() ?? '';
          final dist = row['distanceKm'];
          final label = [
            name,
            if (city.isNotEmpty) city,
            if (dist != null) l10n.nearbyCommercialDistance('$dist'),
          ].join(' — ');
          final selected = _selectedCommercialId == id;
          return ListTile(
            leading: Icon(
              selected ? Icons.radio_button_checked : Icons.radio_button_off,
              color: selected ? AppColors.accent : null,
            ),
            title: Text(label),
            dense: true,
            contentPadding: EdgeInsets.zero,
            onTap: _busy
                ? null
                : () => setState(() => _selectedCommercialId = id),
          );
        }),
        if (_nearby.isNotEmpty)
          TextButton(
            onPressed: _busy ? null : () => setState(() => _selectedCommercialId = null),
            child: Text(l10n.nearbyCommercialSkip),
          ),
      ],
    );
  }

  bool get _hasSocial => _isIOS || GoogleAuth.isConfigured;

  Widget _buildOrDivider(AppLocalizations l10n) {
    return Row(
      children: [
        const Expanded(child: Divider()),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 12),
          child: Text(l10n.loginOr, style: Theme.of(context).textTheme.bodySmall),
        ),
        const Expanded(child: Divider()),
      ],
    );
  }

  List<Widget> _buildSocialButtons(AppLocalizations l10n) {
    return [
      if (_isIOS)
        // Apple HIG: white on dark backgrounds, black on light.
        FilledButton.icon(
          onPressed: _busy ? null : tapApple,
          style: FilledButton.styleFrom(
            backgroundColor: Theme.of(context).brightness == Brightness.dark
                ? Colors.white
                : Colors.black,
            foregroundColor: Theme.of(context).brightness == Brightness.dark
                ? Colors.black
                : Colors.white,
          ),
          icon: const Icon(Icons.apple, size: 24),
          label: Text(l10n.loginWithApple),
        ),
      if (_isIOS && GoogleAuth.isConfigured) const SizedBox(height: 12),
      if (GoogleAuth.isConfigured)
        OutlinedButton.icon(
          onPressed: _busy ? null : submitGoogle,
          icon: const Icon(Icons.g_mobiledata, size: 28),
          label: Text(l10n.loginWithGoogle),
        ),
    ];
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Container(
      decoration: BoxDecoration(gradient: AppTheme.loginGradientOf(context)),
      child: Scaffold(
        backgroundColor: Colors.transparent,
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          title: Text(l10n.registerTitle),
        ),
        body: SafeArea(
          bottom: false,
          child: SingleChildScrollView(
            padding: scrollPaddingWithSystemBottom(context, all: 24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const PetsLogo(height: 28),
                const SizedBox(height: 16),
                Text(l10n.registerSubtitle, textAlign: TextAlign.center),
                const SizedBox(height: 24),
                if (success != null) ...[
                  Text(success!, style: const TextStyle(color: AppColors.accent)),
                  const SizedBox(height: 16),
                  FilledButton(
                    onPressed: () => Navigator.of(context).pop(email.text.trim()),
                    child: Text(l10n.registerBackToLogin),
                  ),
                ] else ...[
                  _buildConsentRow(l10n),
                  if (_hasSocial) ...[
                    const SizedBox(height: 16),
                    ..._buildSocialButtons(l10n),
                  ],
                  if (info != null) ...[
                    const SizedBox(height: 12),
                    Text(info!, style: const TextStyle(color: AppColors.accent)),
                  ],
                  if (error != null) ...[
                    const SizedBox(height: 12),
                    Text(error!, style: const TextStyle(color: AppColors.alert)),
                  ],
                  if (_hasSocial) ...[
                    const SizedBox(height: 16),
                    _buildOrDivider(l10n),
                    const SizedBox(height: 16),
                  ] else if (info != null || error != null) ...[
                    const SizedBox(height: 8),
                  ],
                  TextField(
                    controller: fullName,
                    textInputAction: TextInputAction.next,
                    decoration: InputDecoration(labelText: l10n.fullName),
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    controller: email,
                    keyboardType: TextInputType.emailAddress,
                    textInputAction: TextInputAction.next,
                    decoration: InputDecoration(labelText: l10n.email),
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    controller: password,
                    obscureText: true,
                    textInputAction: TextInputAction.next,
                    decoration: InputDecoration(labelText: l10n.password),
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    controller: confirm,
                    obscureText: true,
                    textInputAction: TextInputAction.next,
                    decoration: InputDecoration(labelText: l10n.confirmNewPassword),
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    key: const Key('register_invite_code'),
                    controller: inviteCodeCtrl,
                    textCapitalization: TextCapitalization.characters,
                    textInputAction: TextInputAction.done,
                    onSubmitted: (_) => submit(),
                    decoration: InputDecoration(
                      labelText: l10n.registerInviteCode,
                      hintText: l10n.registerInviteCodeHint,
                    ),
                  ),
                  _buildNearbySection(l10n),
                  const SizedBox(height: 16),
                  FilledButton(
                    key: const Key('register_submit_btn'),
                    onPressed: _busy ? null : submit,
                    child: Text(l10n.registerSubmit),
                  ),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}
