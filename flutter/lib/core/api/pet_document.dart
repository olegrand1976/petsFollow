import 'dart:io';

import 'package:path_provider/path_provider.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/open_url.dart';

/// Télécharge un document animal (PHI, endpoint authentifié) dans un fichier
/// temporaire puis l'ouvre avec l'application système. Il n'existe pas d'URL
/// publique vers ces objets : le lien direct n'est plus une option.
Future<bool> openPetDocument({
  required String petId,
  required String documentId,
  String? fileName,
}) async {
  final bytes = await ApiClient.instance.downloadPetDocument(petId, documentId);
  if (bytes.isEmpty) return false;
  final dir = await getTemporaryDirectory();
  final file = File('${dir.path}/${_tempFileName(documentId, fileName)}');
  await file.writeAsBytes(bytes, flush: true);
  return openExternalUrl(file.uri.toString());
}

final _extPattern = RegExp(r'^\.[A-Za-z0-9]{1,8}$');

/// Le nom d'origine vient de l'utilisateur : on n'en conserve que l'extension,
/// nécessaire pour que l'OS choisisse la bonne application.
String _tempFileName(String documentId, String? fileName) {
  final name = fileName ?? '';
  final dot = name.lastIndexOf('.');
  final ext = dot > 0 ? name.substring(dot) : '';
  return 'document-$documentId${_extPattern.hasMatch(ext) ? ext : ''}';
}
