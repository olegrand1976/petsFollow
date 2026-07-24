/// Pending email / info banner applied the next time [LoginScreen] builds.
class PendingLoginHint {
  PendingLoginHint._();

  static String? email;
  static String? infoMessage;

  static bool get hasPending =>
      (email != null && email!.isNotEmpty) || (infoMessage != null && infoMessage!.isNotEmpty);

  static void set({String? email, String? infoMessage}) {
    PendingLoginHint.email = email?.trim().isEmpty == true ? null : email?.trim();
    PendingLoginHint.infoMessage = infoMessage;
  }

  static ({String? email, String? infoMessage}) take() {
    final e = email;
    final i = infoMessage;
    email = null;
    infoMessage = null;
    return (email: e, infoMessage: i);
  }
}
