/// Species that support heart-rate measurement.
/// Unknown/`null` is treated as allowed (caller may omit species); only
/// explicit non-HR species such as `other` are blocked.
bool supportsHeartRateControl(String? species) {
  if (species == null || species.isEmpty) return true;
  switch (species) {
    case 'dog':
    case 'cat':
    case 'horse':
      return true;
    default:
      return false;
  }
}
